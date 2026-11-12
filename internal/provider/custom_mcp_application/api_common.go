package custom_mcp_application

import (
	"regexp"
	"unicode/utf8"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	apiResourceDescription = "A custom API app: a REST API described here so `hush_api_application` " +
		"can create applications from it (with `app_catalog_id` set to this resource's " +
		"`app_catalog_id`), each serving its enabled tools to agents. Changes to the address, " +
		"headers, credential and tools carry over to every application created from it. It " +
		"cannot be destroyed while an application still uses it.\n\n" +
		"Detecting the tools from the OpenAPI spec and optimizing them run in the console, not " +
		"here: detect or optimize there, review the tools, then write them as `tool` blocks " +
		"(`terraform import` reads them back, `operation_id` included)."
	apiURLDesc     = "The API's address. A tool's `path` is taken from its host, not appended to its path."
	apiSpecURLDesc = "The address of the API's OpenAPI spec, which the console detects the tools from."
	apiHeadersDesc = "Extra HTTP headers the gateway sends with every request to the API. " +
		"Names the gateway owns (Authorization, Host, Content-Type and the like, and any starting " +
		"with 'proxy-', 'x-hush-' or 'hush-') are refused."
	apiDetectRequiredDesc = "Whether `url` or `spec_url` changed since the tools were last detected."
	apiAuthDesc           = "A fixed credential the gateway presents to the API."
	apiAuthTypeDesc       = "'bearer' (sent as 'Authorization: Bearer <secret>'), 'basic' (sent as " +
		"basic auth with `username`), 'header' (sent as header `name`) or 'client_credentials' (an " +
		"OAuth client, `client_id` with the secret, that the gateway exchanges at `token_url` for " +
		"access tokens it sends as bearer tokens)."
	apiToolDesc = "A tool the API offers: one operation of it. The blocks replace the app's whole " +
		"tool list, which agents get the enabled tools of; with none written, whatever the custom " +
		"app holds (detected in the console, for instance) is left alone. While applications use " +
		"the app, at least one tool must be enabled. The operation the gateway applies to a tool " +
		"is set per application, by `tool_operation`."
	apiToolNameDesc        = "The tool's name, up to 128 characters without whitespace."
	apiToolTitleDesc       = "A short title, up to 64 characters."
	apiToolEnabledDesc     = "Whether agents get the tool. Defaults to true."
	apiToolMethodDesc      = "The HTTP method: 'GET', 'PUT', 'POST', 'DELETE', 'OPTIONS', 'HEAD', 'PATCH' or 'TRACE'."
	apiToolPathDesc        = "The path, from the host of `url`. Its `{placeholders}` must be the path arguments of `input_schema`."
	apiToolInputSchemaDesc = "The tool's arguments as a JSON object schema whose properties are " +
		"any of 'path', 'query', 'header' and 'body', each an object schema of its arguments. " +
		"Use `jsonencode`. Defaults to `{\"type\":\"object\"}`, no arguments."
	apiToolContentTypeDesc = "How the body is sent: 'application/json' (the default) or 'application/octet-stream'."
	apiToolOperationIDDesc = "The spec's operationId of a tool detected in the console, which the " +
		"next detect matches it by, so it can be renamed. Only a tool the app already has can " +
		"name one, so a new app takes none: create it without, detect, then import. Leave unset " +
		"for a tool written by hand."
)

var apiToolNamePattern = regexp.MustCompile(`^\S{1,128}$`)

func APIResourceSchema() map[string]*schema.Schema {
	mcp := ResourceSchema()
	headers := mcp["headers"]
	headers.Description = apiHeadersDesc
	return map[string]*schema.Schema{
		"name":           mcp["name"],
		"app_catalog_id": mcp["app_catalog_id"],
		"display_name":   mcp["display_name"],
		"description":    mcp["description"],
		"url": {
			Description:  apiURLDesc,
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
		},
		"spec_url": {
			Description:  apiSpecURLDesc,
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
		},
		"detect_required": {
			Description: apiDetectRequiredDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"headers": headers,
		"auth":    authSchema(apiAuthDesc, apiAuthTypeDesc),
		"tool": {
			Description: apiToolDesc,
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Description:  apiToolNameDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(apiToolNamePattern, "must be 1 to 128 characters without whitespace"),
				},
				"title": {
					Description:      apiToolTitleDesc,
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: maxChars(64),
				},
				"description": {Description: toolDescDesc, Type: schema.TypeString, Optional: true},
				"type": {
					Description:  toolTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(client.ToolTypes, false),
				},
				"enabled": {Description: apiToolEnabledDesc, Type: schema.TypeBool, Optional: true, Default: true},
				"method": {
					Description:  apiToolMethodDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(client.APIToolMethods, false),
				},
				"path": {
					Description:  apiToolPathDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/[^?#\s]*$`), "must start with '/' and have no query, fragment or whitespace"),
				},
				"input_schema": {
					Description:      apiToolInputSchemaDesc,
					Type:             schema.TypeString,
					Optional:         true,
					Default:          `{"type":"object"}`,
					ValidateFunc:     validation.StringIsJSON,
					DiffSuppressFunc: structure.SuppressJsonDiff,
				},
				"content_type": {
					Description:  apiToolContentTypeDesc,
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "application/json",
					ValidateFunc: validation.StringInSlice(client.APIToolContentTypes, false),
				},
				"operation_id": {Description: apiToolOperationIDDesc, Type: schema.TypeString, Optional: true},
			}},
		},
	}
}

// maxChars counts characters, as heimdall's max_length does, not bytes.
func maxChars(n int) schema.SchemaValidateDiagFunc {
	return func(v any, _ cty.Path) diag.Diagnostics {
		if utf8.RuneCountInString(v.(string)) > n {
			return diag.Errorf("must be at most %d characters", n)
		}
		return nil
	}
}
