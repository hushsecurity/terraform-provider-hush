package notification_channel

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

const (
	idDesc          = "The unique identifier of the notification channel"
	nameDesc        = "The name of the notification channel"
	descriptionDesc = "The description of the notification channel"
	enabledDesc     = "Whether the notification channel is enabled"
)

func NotificationChannelResourceSchema() map[string]*schema.Schema {
	s := NotificationChannelDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description: nameDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["description"] = &schema.Schema{
		Description: descriptionDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["enabled"] = &schema.Schema{
		Description: enabledDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	}

	s["email_config"] = &schema.Schema{
		Description:   "Email notification configuration",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100,
		ConflictsWith: []string{"webhook_config", "slack_config", "splunk_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address": {
					Description: "Email address for notifications",
					Type:        schema.TypeString,
					Required:    true,
				},
				"verified": {
					Description: "Whether the email address is verified",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			},
		},
	}

	s["webhook_config"] = &schema.Schema{
		Description:   "Webhook notification configuration. Multiple webhook_config blocks can be specified to send notifications to multiple webhook URLs.",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100, // API limit as confirmed in analysis
		ConflictsWith: []string{"email_config", "slack_config", "splunk_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Description: "Webhook URL",
					Type:        schema.TypeString,
					Required:    true,
				},
				"method": {
					Description: "HTTP method for webhook requests",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "POST",
					ValidateFunc: validation.StringInSlice([]string{
						string(client.WebhookMethodPOST),
						string(client.WebhookMethodGET),
					}, false),
				},
				"onprem_deployment_id": {
					Description: "Deliver through this on-prem deployment's access bridge instead of directly over the internet. Bridge-bound urls may use http or https on any port and may name an internal host; direct ones must be https on 443 with a public TLD.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"payload_format": {
					Description: "text is the human-readable blob; json is a structured envelope for SIEM/SOAR consumers.",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     string(client.WebhookPayloadFormatText),
					ValidateFunc: validation.StringInSlice([]string{
						string(client.WebhookPayloadFormatText),
						string(client.WebhookPayloadFormatJSON),
					}, false),
				},
				"tls_verify": {
					Description: "Whether to validate the endpoint's TLS certificate. Only bridge-bound endpoints may turn this off, for an internal endpoint served from a private CA.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"auth": {
					Description: "Credential the endpoint requires. The API never returns it, so it is only ever sent.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Description: "bearer sends `Authorization: Bearer <credential>`, basic sends `Authorization: Basic <base64(username:credential)>`, header sends `<name>: <credential>`.",
								Type:        schema.TypeString,
								Required:    true,
								ValidateFunc: validation.StringInSlice([]string{
									string(client.WebhookAuthTypeBearer),
									string(client.WebhookAuthTypeBasic),
									string(client.WebhookAuthTypeHeader),
								}, false),
							},
							"username": {
								Description: "Username, for type basic.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"name": {
								Description: "Header name, for type header. Framing and hop-by-hop headers are rejected.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"credential": {
								Description: "The secret itself -- the bearer token, the basic password, or the header value. Stored in Terraform state; prefer credential_wo.",
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
							},
							"credential_wo": {
								Description: "The secret itself, kept out of Terraform state. Changing it takes effect when credential_wo_version changes.",
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								WriteOnly:   true,
							},
							"credential_wo_version": {
								Description: "Bump to re-send credential_wo. Required with it: a write-only value is not in state, so without a version a rotated credential would never be sent.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
				"verified": {
					Description: "Whether the webhook URL is verified",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			},
		},
	}

	s["slack_config"] = &schema.Schema{
		Description:   "Slack notification configuration. Multiple slack_config blocks can be specified to send notifications to multiple Slack channels.",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100, // API limit as confirmed in analysis
		ConflictsWith: []string{"email_config", "webhook_config", "splunk_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"integration_id": {
					Description: "Slack integration ID",
					Type:        schema.TypeString,
					Required:    true,
				},
				"channel": {
					Description: "Slack channel name",
					Type:        schema.TypeString,
					Required:    true,
				},
				"channel_id": {
					Description: "Slack channel ID",
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}

	s["splunk_config"] = &schema.Schema{
		Description:   "Splunk HTTP Event Collector destination. The collector path, the Authorization scheme and the event envelope are fixed by the API; give it the HEC host and a token. The destination is proven against HEC when the token is saved, so it needs no verification.",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100,
		ConflictsWith: []string{"email_config", "webhook_config", "slack_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Description:      "The HEC host, e.g. `https://http-inputs-acme.splunkcloud.com` or `https://splunk.internal:8088`. The API appends `/services/collector`; any spelling of the collector path is accepted and read back as one. Direct urls must be https on 443; bridge-bound ones may use any port.",
					Type:             schema.TypeString,
					Required:         true,
					DiffSuppressFunc: suppressCollectorPath,
				},
				"token": {
					Description: "The HEC token. Stored in Terraform state; prefer token_wo. The API never returns it.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
				},
				"token_wo": {
					Description: "The HEC token, kept out of Terraform state. Changing it takes effect when token_wo_version changes.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					WriteOnly:   true,
				},
				"token_wo_version": {
					Description: "Bump to re-send token_wo. Required with it: a write-only value is not in state, so without a version a rotated token would never be sent.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"index": {
					Description: "Splunk index to write to. Unset uses the token's default index.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"sourcetype": {
					Description: "Splunk sourcetype of the events.",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     splunkDefaultSourcetype,
				},
				"onprem_deployment_id": {
					Description: "Deliver through this on-prem deployment's access bridge instead of directly over the internet. Most Splunk Enterprise collectors need it.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tls_verify": {
					Description: "Whether to validate HEC's TLS certificate. Only bridge-bound destinations may turn this off, for the self-signed certificate HEC ships with.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"verified": {
					Description: "Always true: the token was accepted by HEC when it was saved.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			},
		},
	}

	return s
}

const (
	splunkCollectorPath     = "/services/collector"
	splunkDefaultSourcetype = "hush:notification"
)

// splunkCollectorURL is the url the API stores for a given HEC host: the bare
// collector, whatever spelling of it the configuration carried.
func splunkCollectorURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	path := strings.TrimRight(parsed.Path, "/")
	if i := strings.Index(path, splunkCollectorPath); i >= 0 {
		path = path[:i]
	}
	parsed.Path = path + splunkCollectorPath
	return parsed.String()
}

// The API answers with the collector path appended, so a configuration that
// names the host alone would otherwise plan a change on every run.
func suppressCollectorPath(_, old, new string, _ *schema.ResourceData) bool {
	return splunkCollectorURL(old) == splunkCollectorURL(new)
}

func NotificationChannelDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description:   idDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"id"},
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enabled": {
			Description: enabledDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"type": {
			Description: "The type of notification channel",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"email_config": {
			Description: "Email notification configuration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"address": {
						Description: "Email address for notifications",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"verified": {
						Description: "Whether the email address is verified",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
		"webhook_config": {
			Description: "Webhook notification configuration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Description: "Webhook URL",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"method": {
						Description: "HTTP method for webhook requests",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"onprem_deployment_id": {
						Description: "Deployment whose access bridge delivers this webhook",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"payload_format": {
						Description: "text or json",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"tls_verify": {
						Description: "Whether the endpoint's TLS certificate is validated",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"auth": {
						Description: "Credential shape the endpoint is configured with. The credential itself is never returned.",
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"type":     {Type: schema.TypeString, Computed: true},
								"username": {Type: schema.TypeString, Computed: true},
								"name":     {Type: schema.TypeString, Computed: true},
							},
						},
					},
					"verified": {
						Description: "Whether the webhook URL is verified",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
		"slack_config": {
			Description: "Slack notification configuration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"integration_id": {
						Description: "Slack integration ID",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"channel": {
						Description: "Slack channel name",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"channel_id": {
						Description: "Slack channel ID",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"splunk_config": {
			Description: "Splunk HTTP Event Collector destination. The token is never returned.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Description: "The collector url",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"index": {
						Description: "Splunk index, empty for the token's default",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"sourcetype": {
						Description: "Splunk sourcetype of the events",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"onprem_deployment_id": {
						Description: "Deployment whose access bridge delivers to the collector",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"tls_verify": {
						Description: "Whether HEC's TLS certificate is validated",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"verified": {
						Description: "Always true: the token was accepted by HEC when it was saved",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
	}
}

// Helper Functions

func notificationChannelRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var channel *client.NotificationChannel
	var err error

	if id := d.Id(); id != "" {
		channel, err = client.GetNotificationChannel(ctx, c, id)
		if err != nil {
			// Handle 404 errors gracefully by removing from state
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			} else {
				return diag.FromErr(err)
			}
		}
	} else if id, exists := d.GetOk("id"); exists {
		channelID := id.(string)
		channel, err = client.GetNotificationChannel(ctx, c, channelID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no notification channel found with ID: %s", channelID)
			} else {
				return diag.FromErr(err)
			}
		}
	} else if name, exists := d.GetOk("name"); exists {
		channelName := name.(string)
		channels, err := client.GetNotificationChannelsByName(ctx, c, channelName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup notification channel by name '%s': %w", channelName, err))
		}

		switch len(channels) {
		case 0:
			return diag.Errorf("no notification channel found with name: %s", channelName)
		case 1:
			channel = &channels[0]
		default:
			return diag.Errorf("multiple notification channels found with name '%s'. Channel names must be unique. Consider using the channel ID instead for exact matching", channelName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(channel.ID)
	}

	if diags := setNotificationChannelFields(d, channel); diags.HasError() {
		return diags
	}

	return nil
}

func setNotificationChannelFields(d *schema.ResourceData, channel *client.NotificationChannel) diag.Diagnostics {
	fields := map[string]any{
		"name":        channel.Name,
		"description": channel.Description,
		"enabled":     channel.Enabled,
		"type":        string(channel.Type),
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

type webhookCredential struct{ credential, version string }

// nilIfEmpty distinguishes "not configured" from "not asked about": the api
// reads an absent key as the latter.
func nilIfEmpty(value any) any {
	if s, ok := value.(string); ok && s == "" {
		return nil
	}
	return value
}

func urlOf(config map[string]any) string {
	url, _ := config["url"].(string)
	return url
}

// valueOr keeps a field the api omitted at its schema default, rather than
// letting d.Set store a zero that plans as drift.
func valueOr[T any](value any, fallback T) any {
	if value == nil {
		return fallback
	}
	if s, ok := value.(string); ok && s == "" {
		return fallback
	}
	return value
}

// The api never returns a credential, so a refresh keeps whatever the
// configuration holds rather than blanking it and provoking a diff. Keyed by
// url so a reordered response does not move one endpoint's credential onto
// another.
func existingWebhookCredentials(d *schema.ResourceData) map[string]webhookCredential {
	out := map[string]webhookCredential{}
	configs, ok := d.Get("webhook_config").([]any)
	if !ok {
		return out
	}
	for _, configInterface := range configs {
		configMap, ok := configInterface.(map[string]any)
		if !ok {
			continue
		}
		blocks, _ := configMap["auth"].([]any)
		if len(blocks) == 0 || blocks[0] == nil {
			continue
		}
		authMap, _ := blocks[0].(map[string]any)
		credential, _ := authMap["credential"].(string)
		version, _ := authMap["credential_wo_version"].(string)
		out[urlOf(configMap)] = webhookCredential{credential, version}
	}
	return out
}

// existingSplunkTokens is the Splunk counterpart: keyed by the normalized
// collector url, since the configuration may name the host alone while the
// api answers with the path appended.
func existingSplunkTokens(d *schema.ResourceData) map[string]webhookCredential {
	out := map[string]webhookCredential{}
	configs, ok := d.Get("splunk_config").([]any)
	if !ok {
		return out
	}
	for _, configInterface := range configs {
		configMap, ok := configInterface.(map[string]any)
		if !ok {
			continue
		}
		token, _ := configMap["token"].(string)
		version, _ := configMap["token_wo_version"].(string)
		out[splunkCollectorURL(urlOf(configMap))] = webhookCredential{token, version}
	}
	return out
}

func setNotificationChannelConfigFields(d *schema.ResourceData, channel *client.NotificationChannel) error {
	// Captured before the clear below, since that is what holds them.
	configuredCredentials := existingWebhookCredentials(d)
	configuredTokens := existingSplunkTokens(d)

	if err := d.Set("email_config", nil); err != nil {
		return fmt.Errorf("failed to clear email_config: %w", err)
	}
	if err := d.Set("webhook_config", nil); err != nil {
		return fmt.Errorf("failed to clear webhook_config: %w", err)
	}
	if err := d.Set("slack_config", nil); err != nil {
		return fmt.Errorf("failed to clear slack_config: %w", err)
	}
	if err := d.Set("splunk_config", nil); err != nil {
		return fmt.Errorf("failed to clear splunk_config: %w", err)
	}

	switch channel.Type {
	case client.NotificationChannelTypeEmail:
		if len(channel.Config) > 0 {
			emailConfigs := make([]map[string]any, len(channel.Config))
			for i, config := range channel.Config {
				emailConfigs[i] = map[string]any{
					"address":  config["address"],
					"verified": config["verified"],
				}
			}
			if err := d.Set("email_config", emailConfigs); err != nil {
				return fmt.Errorf("failed to set email_config: %w", err)
			}
		}
	case client.NotificationChannelTypeWebhook:
		if len(channel.Config) > 0 {
			webhookConfigs := make([]map[string]any, len(channel.Config))
			for i, config := range channel.Config {
				webhookConfigs[i] = map[string]any{
					"url":                  config["url"],
					"method":               config["method"],
					"onprem_deployment_id": config["onprem_deployment_id"],
					// Defaulted, or a channel stored before these existed plans
					// a change on every run and never converges.
					"payload_format": valueOr(config["payload_format"], string(client.WebhookPayloadFormatText)),
					"tls_verify":     valueOr(config["tls_verify"], true),
					"verified":       config["verified"],
				}
				if auth, ok := config["auth"].(map[string]any); ok {
					authConfig := map[string]any{
						"type":     auth["type"],
						"username": auth["username"],
						"name":     auth["name"],
					}
					// The credential is never returned, so the configured one
					// is carried across rather than refreshed away. Keyed by
					// url, since the api need not answer in the order it was
					// sent. Absent keys are left out: the data source schema
					// has no credential to set them on.
					if credential, ok := configuredCredentials[urlOf(config)]; ok {
						if credential.credential != "" {
							authConfig["credential"] = credential.credential
						}
						if credential.version != "" {
							authConfig["credential_wo_version"] = credential.version
						}
					}
					webhookConfigs[i]["auth"] = []map[string]any{authConfig}
				}
			}
			if err := d.Set("webhook_config", webhookConfigs); err != nil {
				return fmt.Errorf("failed to set webhook_config: %w", err)
			}
		}
	case client.NotificationChannelTypeSlack:
		if len(channel.Config) > 0 {
			slackConfigs := make([]map[string]any, len(channel.Config))
			for i, config := range channel.Config {
				slackConfigs[i] = map[string]any{
					"integration_id": config["integration_id"],
					"channel":        config["channel"],
					"channel_id":     config["channel_id"],
				}
			}
			if err := d.Set("slack_config", slackConfigs); err != nil {
				return fmt.Errorf("failed to set slack_config: %w", err)
			}
		}
	case client.NotificationChannelTypeSplunk:
		if len(channel.Config) > 0 {
			splunkConfigs := make([]map[string]any, len(channel.Config))
			for i, config := range channel.Config {
				splunkConfigs[i] = map[string]any{
					"url":                  config["url"],
					"index":                config["index"],
					"sourcetype":           valueOr(config["sourcetype"], splunkDefaultSourcetype),
					"onprem_deployment_id": config["onprem_deployment_id"],
					"tls_verify":           valueOr(config["tls_verify"], true),
					"verified":             config["verified"],
				}
				// Never returned, so carried across like a webhook credential.
				// Absent keys are left out: the data source has none to set.
				if token, ok := configuredTokens[splunkCollectorURL(urlOf(config))]; ok {
					if token.credential != "" {
						splunkConfigs[i]["token"] = token.credential
					}
					if token.version != "" {
						splunkConfigs[i]["token_wo_version"] = token.version
					}
				}
			}
			if err := d.Set("splunk_config", splunkConfigs); err != nil {
				return fmt.Errorf("failed to set splunk_config: %w", err)
			}
		}
	}

	return nil
}

// ValidateSplunkToken is the token's counterpart of ValidateWebhookAuth: the
// pairing rules live inside a list block, out of the schema's reach.
func ValidateSplunkToken(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	configs, ok := diff.Get("splunk_config").([]any)
	if !ok {
		return nil
	}
	for i := range configs {
		set := func(attr string) bool {
			return writeonly.IsSetNested(diff, "splunk_config", i, attr)
		}
		switch {
		case set("token") && set("token_wo"):
			return fmt.Errorf("splunk_config[%d]: token and token_wo are mutually exclusive", i)
		case !set("token") && !set("token_wo"):
			return fmt.Errorf("splunk_config[%d]: one of token or token_wo is required", i)
		case set("token_wo") && !set("token_wo_version"):
			return fmt.Errorf("splunk_config[%d]: token_wo requires token_wo_version", i)
		}
	}
	return nil
}

// ValidateWebhookAuth rejects auth blocks the schema cannot check itself:
// ConflictsWith and RequiredWith address top-level attributes only, and auth
// lives inside a list block. Everything here reads raw config rather than
// diff.Get, so a value Terraform resolves at apply counts as configured
// instead of aborting the plan.
func ValidateWebhookAuth(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	configs, ok := diff.Get("webhook_config").([]any)
	if !ok {
		return nil
	}
	for i, configInterface := range configs {
		configMap, ok := configInterface.(map[string]any)
		if !ok {
			continue
		}
		blocks, _ := configMap["auth"].([]any)
		if len(blocks) == 0 || blocks[0] == nil {
			continue
		}
		authMap, _ := blocks[0].(map[string]any)
		authPath := []any{"webhook_config", i, "auth", 0}
		set := func(attr string) bool {
			return writeonly.IsSetNested(diff, append(authPath, attr)...)
		}

		switch {
		case set("credential") && set("credential_wo"):
			return fmt.Errorf("webhook_config[%d].auth: credential and credential_wo are mutually exclusive", i)
		case !set("credential") && !set("credential_wo"):
			return fmt.Errorf("webhook_config[%d].auth: one of credential or credential_wo is required", i)
		// Without a version there is nothing in state to change, so a rotated
		// write-only credential would never be sent and never be noticed.
		case set("credential_wo") && !set("credential_wo_version"):
			return fmt.Errorf("webhook_config[%d].auth: credential_wo requires credential_wo_version", i)
		}

		// Checked here rather than only on send, so the plan fails instead of
		// the apply.
		authType, _ := authMap["type"].(string)
		username, _ := authMap["username"].(string)
		name, _ := authMap["name"].(string)
		if client.WebhookAuthType(authType) == client.WebhookAuthTypeBasic && username == "" {
			return fmt.Errorf("webhook_config[%d].auth: username is required for type basic", i)
		}
		if client.WebhookAuthType(authType) == client.WebhookAuthTypeHeader && name == "" {
			return fmt.Errorf("webhook_config[%d].auth: name is required for type header", i)
		}
	}
	return nil
}

// webhookAuth builds the auth object for one endpoint. The credential is read
// from raw config when it is write-only, since those never reach state.
func webhookAuth(d *schema.ResourceData, i int, configMap map[string]any) (map[string]any, error) {
	blocks, _ := configMap["auth"].([]any)
	if len(blocks) == 0 || blocks[0] == nil {
		return nil, nil
	}
	authMap, _ := blocks[0].(map[string]any)
	authType, _ := authMap["type"].(string)

	credential, _ := authMap["credential"].(string)
	if credential == "" {
		credential = writeonly.GetNestedString(d, "webhook_config", i, "auth", 0, "credential_wo")
	}
	if credential == "" {
		return nil, fmt.Errorf("webhook_config[%d].auth: one of credential or credential_wo is required", i)
	}

	auth := map[string]any{"type": authType}
	switch client.WebhookAuthType(authType) {
	case client.WebhookAuthTypeBearer:
		auth["token"] = credential
	case client.WebhookAuthTypeBasic:
		username, _ := authMap["username"].(string)
		if username == "" {
			return nil, fmt.Errorf("webhook_config[%d].auth: username is required for type basic", i)
		}
		auth["username"] = username
		auth["password"] = credential
	case client.WebhookAuthTypeHeader:
		name, _ := authMap["name"].(string)
		if name == "" {
			return nil, fmt.Errorf("webhook_config[%d].auth: name is required for type header", i)
		}
		auth["name"] = name
		auth["value"] = credential
	default:
		// Or a new type would send a request that authenticates with nothing.
		return nil, fmt.Errorf("webhook_config[%d].auth: unsupported type %q", i, authType)
	}
	return auth, nil
}

func getNotificationChannelTypeAndConfig(d *schema.ResourceData) (client.NotificationChannelType, []map[string]any, error) {
	if emailConfigs, ok := d.GetOk("email_config"); ok {
		configList := emailConfigs.([]any)
		if len(configList) > 0 {
			result := make([]map[string]any, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]any)
				result[i] = map[string]any{
					"address": configMap["address"],
				}
			}
			return client.NotificationChannelTypeEmail, result, nil
		}
	}

	if webhookConfigs, ok := d.GetOk("webhook_config"); ok {
		configList := webhookConfigs.([]any)
		if len(configList) > 0 {
			result := make([]map[string]any, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]any)
				result[i] = map[string]any{
					"url":        configMap["url"],
					"tls_verify": configMap["tls_verify"],
				}
				if method, ok := configMap["method"]; ok && method != "" {
					result[i]["method"] = method
				}
				if format, ok := configMap["payload_format"]; ok && format != "" {
					result[i]["payload_format"] = format
				}
				// Always sent, null included: an absent key means "inherit
				// what is stored" to the api, so omitting it would both lose
				// the binding on a url change and make removal impossible.
				result[i]["onprem_deployment_id"] = nilIfEmpty(configMap["onprem_deployment_id"])

				auth, err := webhookAuth(d, i, configMap)
				if err != nil {
					return "", nil, err
				}
				// Same: an explicit null is what clears a stored credential.
				// Assigned through a nil check, or a typed nil map would
				// marshal as {} and be rejected for having no type.
				if auth == nil {
					result[i]["auth"] = nil
				} else {
					result[i]["auth"] = auth
				}
			}
			return client.NotificationChannelTypeWebhook, result, nil
		}
	}

	if slackConfigs, ok := d.GetOk("slack_config"); ok {
		configList := slackConfigs.([]any)
		if len(configList) > 0 {
			result := make([]map[string]any, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]any)
				result[i] = map[string]any{
					"integration_id": configMap["integration_id"],
					"channel":        configMap["channel"],
				}
			}
			return client.NotificationChannelTypeSlack, result, nil
		}
	}

	if splunkConfigs, ok := d.GetOk("splunk_config"); ok {
		configList := splunkConfigs.([]any)
		if len(configList) > 0 {
			result := make([]map[string]any, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]any)
				token, _ := configMap["token"].(string)
				if token == "" {
					token = writeonly.GetNestedString(d, "splunk_config", i, "token_wo")
				}
				if token == "" {
					return "", nil, fmt.Errorf("splunk_config[%d]: one of token or token_wo is required", i)
				}
				result[i] = map[string]any{
					// Named: a Splunk item that carries only fields a webhook
					// also has would otherwise be read as one.
					"type":       string(client.NotificationChannelTypeSplunk),
					"url":        configMap["url"],
					"token":      token,
					"sourcetype": configMap["sourcetype"],
					"tls_verify": configMap["tls_verify"],
					// Always sent, null included: an absent key means "leave
					// what is stored" to the api, so neither could be removed.
					"index":                nilIfEmpty(configMap["index"]),
					"onprem_deployment_id": nilIfEmpty(configMap["onprem_deployment_id"]),
				}
			}
			return client.NotificationChannelTypeSplunk, result, nil
		}
	}

	return "", nil, fmt.Errorf("exactly one of email_config, webhook_config, slack_config, or splunk_config must be specified")
}
