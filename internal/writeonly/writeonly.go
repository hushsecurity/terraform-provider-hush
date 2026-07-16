// Package writeonly reads Terraform write-only attribute values.
//
// Write-only attributes (schema.Schema{WriteOnly: true}) are not persisted
// to state, so they are not accessible via *schema.ResourceData's Get/GetOk —
// those calls always return the zero value. The values are only present in
// the raw config (d.GetRawConfig()).
//
// Use GetString in resource Create/Update to extract a secret whose value
// may come from either a plain attribute or a write-only counterpart. Use
// IsSet in CustomizeDiff to check whether a write-only attribute is set.
package writeonly

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// rawConfigGetter is implemented by both *schema.ResourceData and *schema.ResourceDiff.
type rawConfigGetter interface {
	GetRawConfig() cty.Value
}

// GetString returns the value of plainAttr if set, otherwise the value of the
// write-only attribute woAttr read from raw config. Returns "" if neither
// is set.
func GetString(d *schema.ResourceData, plainAttr, woAttr string) string {
	if v, ok := d.GetOk(plainAttr); ok {
		return v.(string)
	}
	return readRawString(d, woAttr)
}

// IsSet reports whether the string attribute attr is configured in raw config.
// A present-but-unknown value (a reference resolved at apply) counts as set;
// null and known-empty do not. Works for plain and write-only attributes.
func IsSet(g rawConfigGetter, attr string) bool {
	rc := g.GetRawConfig()
	if rc.IsNull() {
		return false
	}
	v := rc.GetAttr(attr)
	if v.IsNull() {
		return false
	}
	if !v.IsKnown() {
		return true
	}
	return v.AsString() != ""
}

func readRawString(g rawConfigGetter, attr string) string {
	rc := g.GetRawConfig()
	if rc.IsNull() {
		return ""
	}
	v := rc.GetAttr(attr)
	if v.IsNull() || !v.IsKnown() {
		return ""
	}
	return v.AsString()
}
