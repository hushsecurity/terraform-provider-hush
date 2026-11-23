package cloudflare_r2_access_privilege

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc           = "The unique identifier of the Cloudflare R2 access privilege"
	nameDesc         = "The name of the Cloudflare R2 access privilege"
	descriptionDesc  = "The description of the Cloudflare R2 access privilege"
	permissionDesc   = "The access level of the R2 tokens, as the Cloudflare dashboard names it: `object_read_write` or `object_read_only` reach the objects of `buckets`, or of every bucket with `all_buckets`; `admin_read_write` or `admin_read_only` are account wide, and `admin_read_write` can also manage buckets"
	bucketsDesc      = "The buckets an object permission reaches: 1 to 1000 R2 bucket names. Names are not checked against the account, so a misspelled one, or one in another jurisdiction, gives credentials that S3 refuses for it. Not valid with an admin permission"
	allBucketsDesc   = "Grant an object permission every bucket of the account, including buckets created later. An object permission needs `buckets` or `all_buckets = true`, so the widest grant never comes from leaving `buckets` out. Not valid with an admin permission"
	jurisdictionDesc = "The jurisdiction of the buckets: `default`, `eu`, `fedramp` or `us`. A bucket in a jurisdiction answers only on that jurisdiction's S3 endpoint, so one privilege covers one jurisdiction. Defaults to `default`"
	typeDesc         = "The type of access privilege"
)

// Bounds the token policy the access manager sends, as the Hush API does.
const maxBuckets = 1000

var (
	permissions      = []string{"object_read_write", "object_read_only", "admin_read_write", "admin_read_only"}
	adminPermissions = map[string]bool{"admin_read_write": true, "admin_read_only": true}
	jurisdictions    = []string{"default", "eu", "fedramp", "us"}
	// R2's bucket naming rules, as the Hush API checks them: a "*" in a name
	// would widen what the token reaches.
	bucketRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)
)

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
	s["permission"] = &schema.Schema{
		Description:  permissionDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(permissions, false),
	}
	// A set, as the Hush API sorts the buckets: a list would show any other
	// order as a change.
	s["buckets"] = &schema.Schema{
		Description: bucketsDesc,
		Type:        schema.TypeSet,
		Optional:    true,
		MaxItems:    maxBuckets,
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateFunc: validation.StringMatch(bucketRe,
				"a bucket name must be 3-63 lowercase letters, digits or '-', with no '-' at either end"),
		},
	}
	s["all_buckets"] = &schema.Schema{
		Description: allBucketsDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	}
	s["jurisdiction"] = &schema.Schema{
		Description:  jurisdictionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "default",
		ValidateFunc: validation.StringInSlice(jurisdictions, false),
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
		"permission": {
			Description: permissionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"buckets": {
			Description: bucketsDesc,
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"all_buckets": {
			Description: allBucketsDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"jurisdiction": {
			Description: jurisdictionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func validateBuckets(_ context.Context, d *schema.ResourceDiff, _ any) error {
	c := d.GetRawConfig()
	return checkBuckets(c.GetAttr("permission"), c.GetAttr("buckets"), c.GetAttr("all_buckets"))
}

// checkBuckets mirrors the Hush API on the raw config, which tells an empty
// set from an omitted all_buckets. Unknown values are left to apply.
func checkBuckets(permission, buckets, allBuckets cty.Value) error {
	if !permission.IsKnown() || permission.IsNull() || !buckets.IsKnown() || !allBuckets.IsKnown() {
		return nil
	}
	hasBuckets := !buckets.IsNull()
	if hasBuckets && buckets.IsWhollyKnown() && buckets.LengthInt() == 0 {
		return errors.New("buckets cannot be empty; set all_buckets = true for every bucket")
	}
	all := !allBuckets.IsNull() && allBuckets.True()
	p := permission.AsString()
	if adminPermissions[p] {
		if hasBuckets || all {
			return fmt.Errorf("%s is account wide and takes no buckets; remove buckets and all_buckets", p)
		}
		return nil
	}
	if hasBuckets && all {
		return errors.New("set buckets or all_buckets = true, not both")
	}
	if !hasBuckets && !all {
		return fmt.Errorf("%s needs buckets, or all_buckets = true for every bucket of the account", p)
	}
	return nil
}

// requestBuckets is what the Hush API takes: the buckets, or nil (null) for
// every bucket and for an admin permission, which the plan leaves without any.
func requestBuckets(d *schema.ResourceData) []string {
	set := d.Get("buckets").(*schema.Set)
	if set.Len() == 0 {
		return nil
	}
	out := make([]string, 0, set.Len())
	for _, v := range set.List() {
		out = append(out, v.(string))
	}
	return out
}
