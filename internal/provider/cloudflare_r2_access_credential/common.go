package cloudflare_r2_access_credential

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the Cloudflare R2 access credential"
	nameDesc          = "The name of the Cloudflare R2 access credential"
	descriptionDesc   = "The description of the Cloudflare R2 access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	accountIDDesc     = "The Cloudflare account ID: 32 lowercase hex characters. Changing it requires a new `api_token` (or a bumped `api_token_wo_version`)"
	apiTokenDesc      = "The root account API token, with the `Account API Tokens` `Edit` permission on the account. The access manager creates and revokes the workloads' R2 tokens with it; it never reaches a workload. Every change issues new credentials to every policy that uses the credential"
	apiTokenWODesc    = "The root account API token (write-only). More secure than `api_token` because Terraform will not store this value in the state file."
	apiTokenWOVerDesc = "Used to trigger updates for `api_token_wo`. Change when the token changes."
	typeDesc          = "The type of access credential"
	kindDesc          = "The kind of access credential"
	secretStoreIDDesc = "The ID of the secret store where this credential is saved (optional)"
)

// validateAPIToken mirrors the Hush API: the access manager never retries a
// token it refuses, and a pasted token often ends in a newline.
func validateAPIToken(v any, k string) ([]string, []error) {
	for _, r := range v.(string) {
		if unicode.IsSpace(r) || !unicode.IsPrint(r) {
			return nil, []error{fmt.Errorf("%s cannot contain whitespace or control characters", k)}
		}
	}
	return nil, nil
}

// diffChecker is the subset of *schema.ResourceDiff validateAccountChange
// needs, so the rule is unit-testable.
type diffChecker interface {
	Id() string
	HasChange(string) bool
	NewValueKnown(string) bool
}

func validateAccountChangeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	return validateAccountChange(d)
}

// validateAccountChange mirrors the Hush API: a root token works only in its
// own account. An unknown account_id is left to the API.
func validateAccountChange(d diffChecker) error {
	if d.Id() == "" || !d.NewValueKnown("account_id") || !d.HasChange("account_id") {
		return nil
	}
	if d.HasChange("api_token") || d.HasChange("api_token_wo_version") {
		return nil
	}
	return errors.New(
		"changing account_id requires a new api_token (or api_token_wo with a bumped api_token_wo_version)")
}

func ResourceSchema() map[string]*schema.Schema {
	s := DataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 1000),
	}
	s["deployment_ids"] = &schema.Schema{
		Description: deploymentIDsDesc + ". Changing this after creation is not supported; the credential must be deleted and recreated.",
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		MaxItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["secret_store_id"] = &schema.Schema{
		Description:  secretStoreIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^sst-`), "secret_store_id must start with 'sst-'"),
	}
	s["account_id"] = &schema.Schema{
		Description: accountIDDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringMatch(
			regexp.MustCompile(`^[0-9a-f]{32}$`),
			"account_id must be 32 lowercase hex characters"),
	}
	s["api_token"] = &schema.Schema{
		Description:   apiTokenDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"api_token_wo"},
		ExactlyOneOf:  []string{"api_token", "api_token_wo"},
		ValidateFunc:  validateAPIToken,
	}
	s["api_token_wo"] = &schema.Schema{
		Description:   apiTokenWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"api_token"},
		ExactlyOneOf:  []string{"api_token", "api_token_wo"},
		RequiredWith:  []string{"api_token_wo_version"},
		ValidateFunc:  validateAPIToken,
	}
	s["api_token_wo_version"] = &schema.Schema{
		Description:  apiTokenWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"api_token_wo"},
	}

	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: idDesc,
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: nameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"deployment_ids": {
			Description: deploymentIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"account_id": {
			Description: accountIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"secret_store_id": {
			Description: secretStoreIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
