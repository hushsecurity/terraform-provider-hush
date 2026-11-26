package cloudinary_access_privilege

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc              = "The unique identifier of the Cloudinary access privilege"
	nameDesc            = "The name of the Cloudinary access privilege"
	descriptionDesc     = "The description of the Cloudinary access privilege"
	cloudNameDesc       = "The cloud name of the product environment the keys are created in: 2 to 128 letters, digits or '-', starting with a letter. Cloudinary matches it case-insensitively"
	roleIDDesc          = "A Cloudinary built-in product environment role for the keys, by its ID, such as `cld::role::prodenv::editor`. Set exactly one of `role_id` and `system_policy_ids`. Master Admin, Admin, Tech Admin and MediaFlows Admin can manage API keys: a workload given one can create keys that Hush never revokes"
	systemPolicyIDsDesc = "Cloudinary system policies for a role the access manager creates for each privilege version: 1 to 200 `cld::policy::global::...` IDs, as Cloudinary's role catalog names them. Set exactly one of `role_id` and `system_policy_ids`. With `cld::policy::global::api_keys::manage` a workload can create keys that Hush never revokes, and with `cld::policy::global::api_keys::view` it can read every key's secret"
	typeDesc            = "The type of access privilege"
)

// Bounds the role the access manager creates, as the Hush API does.
const maxPolicies = 200

const oneGrant = "set exactly one of role_id and system_policy_ids"

var (
	cloudNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]{1,127}$`)
	// Content and saved search policies need a folder or collection.
	policyIDRe = regexp.MustCompile(`^cld::policy::global::[a-z0-9_:]+$`)
)

// validatePolicyID mirrors the Hush API: a global system policy.
func validatePolicyID(v any, k string) ([]string, []error) {
	id := v.(string)
	if !policyIDRe.MatchString(id) {
		return nil, []error{fmt.Errorf("%s: %q is not a cld::policy::global:: policy", k, id)}
	}
	return nil, nil
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
	s["cloud_name"] = &schema.Schema{
		Description: cloudNameDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringMatch(cloudNameRe,
			"cloud_name must be 2 to 128 letters, digits or '-', starting with a letter"),
	}
	s["role_id"] = &schema.Schema{
		Description:  roleIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringIsNotEmpty,
	}
	// A set, as the Hush API sorts the policies: a list would show any other
	// order as a change.
	s["system_policy_ids"] = &schema.Schema{
		Description: systemPolicyIDsDesc,
		Type:        schema.TypeSet,
		Optional:    true,
		MaxItems:    maxPolicies,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validatePolicyID,
		},
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
		"cloud_name": {
			Description: cloudNameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role_id": {
			Description: roleIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"system_policy_ids": {
			Description: systemPolicyIDsDesc,
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func validateGrant(_ context.Context, d *schema.ResourceDiff, _ any) error {
	c := d.GetRawConfig()
	return checkGrant(c.GetAttr("role_id"), c.GetAttr("system_policy_ids"))
}

// checkGrant mirrors the Hush API on the raw config, which tells an empty set
// from an omitted one. Unknown values are left to apply.
func checkGrant(roleID, policies cty.Value) error {
	if !roleID.IsKnown() || !policies.IsKnown() {
		return nil
	}
	if !policies.IsNull() && policies.IsWhollyKnown() && policies.LengthInt() == 0 {
		return errors.New("system_policy_ids cannot be empty; " + oneGrant)
	}
	if roleID.IsNull() == policies.IsNull() {
		return errors.New(oneGrant)
	}
	return nil
}

// requestPolicies is the sorted policies, or nil (null) when a role is set.
func requestPolicies(d *schema.ResourceData) []string {
	set := d.Get("system_policy_ids").(*schema.Set)
	if set.Len() == 0 {
		return nil
	}
	out := make([]string, 0, set.Len())
	for _, v := range set.List() {
		out = append(out, v.(string))
	}
	sort.Strings(out)
	return out
}
