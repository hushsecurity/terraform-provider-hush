package cloudinary_access_credential

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc             = "The unique identifier of the Cloudinary access credential"
	nameDesc           = "The name of the Cloudinary access credential"
	descriptionDesc    = "The description of the Cloudinary access credential"
	deploymentIDsDesc  = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	accountIDDesc      = "The Cloudinary account ID, as **Settings > Account Management Keys** shows it: 1 to 128 letters, digits, '_' or '-'. Changing it requires a new `api_key` and a new `api_secret` (or bumped `api_key_wo_version` and `api_secret_wo_version`)"
	apiKeyDesc         = "The account management key: the key of an Enterprise account's **Settings > Account Management Keys** pair. The access manager creates and deletes the workloads' product environment keys and their roles with it; it never reaches a workload. Every change issues new credentials to every policy that uses the credential"
	apiKeyWODesc       = "The account management key (write-only). More secure than `api_key` because Terraform will not store this value in the state file."
	apiKeyWOVerDesc    = "Used to trigger updates for `api_key_wo`. Change when the key changes."
	apiSecretDesc      = "The secret of the account management key. Every change issues new credentials to every policy that uses the credential"
	apiSecretWODesc    = "The secret of the account management key (write-only). More secure than `api_secret` because Terraform will not store this value in the state file."
	apiSecretWOVerDesc = "Used to trigger updates for `api_secret_wo`. Change when the secret changes."
	typeDesc           = "The type of access credential"
	kindDesc           = "The kind of access credential"
	secretStoreIDDesc  = "The ID of the secret store where this credential is saved (optional)"
)

// The access manager puts the account in API paths, so the Hush API takes no
// other characters.
var accountIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// validateKeyPart mirrors the Hush API: 1 to 256 characters, none of them
// whitespace or control, which the access manager would refuse for good.
func validateKeyPart(v any, k string) ([]string, []error) {
	s := v.(string)
	if n := utf8.RuneCountInString(s); n < 1 || n > 256 {
		return nil, []error{fmt.Errorf("%s must be 1 to 256 characters", k)}
	}
	for _, r := range s {
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

// validateAccountChange mirrors the Hush API: an account management key works
// only in its own account. An unknown account_id is left to the API.
func validateAccountChange(d diffChecker) error {
	if d.Id() == "" || !d.NewValueKnown("account_id") || !d.HasChange("account_id") {
		return nil
	}
	newKey := d.HasChange("api_key") || d.HasChange("api_key_wo_version")
	newSecret := d.HasChange("api_secret") || d.HasChange("api_secret_wo_version")
	if newKey && newSecret {
		return nil
	}
	return errors.New(
		"changing account_id requires a new api_key and a new api_secret (or api_key_wo and api_secret_wo with bumped versions)")
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
		ValidateFunc: validation.StringMatch(accountIDRe,
			"account_id must be 1 to 128 letters, digits, '_' or '-'"),
	}
	s["api_key"] = &schema.Schema{
		Description:   apiKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"api_key_wo"},
		ExactlyOneOf:  []string{"api_key", "api_key_wo"},
		ValidateFunc:  validateKeyPart,
	}
	s["api_key_wo"] = &schema.Schema{
		Description:   apiKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"api_key"},
		ExactlyOneOf:  []string{"api_key", "api_key_wo"},
		RequiredWith:  []string{"api_key_wo_version"},
		ValidateFunc:  validateKeyPart,
	}
	s["api_key_wo_version"] = &schema.Schema{
		Description:  apiKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"api_key_wo"},
	}
	s["api_secret"] = &schema.Schema{
		Description:   apiSecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"api_secret_wo"},
		ExactlyOneOf:  []string{"api_secret", "api_secret_wo"},
		ValidateFunc:  validateKeyPart,
	}
	s["api_secret_wo"] = &schema.Schema{
		Description:   apiSecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"api_secret"},
		ExactlyOneOf:  []string{"api_secret", "api_secret_wo"},
		RequiredWith:  []string{"api_secret_wo_version"},
		ValidateFunc:  validateKeyPart,
	}
	s["api_secret_wo_version"] = &schema.Schema{
		Description:  apiSecretWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"api_secret_wo"},
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
