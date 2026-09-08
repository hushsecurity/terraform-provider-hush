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

// IsSetNested reports whether a nested attribute is configured, treating a
// present-but-unknown value -- a reference Terraform resolves at apply -- as
// set. GetNestedString cannot answer this: it returns "" for unknown, which a
// caller would otherwise read as absent and reject.
func IsSetNested(g rawConfigGetter, path ...any) bool {
	v, ok := walk(g.GetRawConfig(), path)
	if !ok {
		return false
	}
	if !v.IsKnown() {
		return true
	}
	if v.IsNull() || v.Type() != cty.String {
		return false
	}
	return v.AsString() != ""
}

// GetNestedString returns the value of a write-only attribute inside nested
// list blocks. The path alternates block names and list indices, ending in the
// attribute name -- GetNestedString(d, "webhook_config", 0, "auth", 0, "credential_wo")
// reads the credential of the first webhook endpoint's auth block. Returns ""
// if any step is missing, null or not yet known.
func GetNestedString(g rawConfigGetter, path ...any) string {
	v, ok := walk(g.GetRawConfig(), path)
	if !ok || v.IsNull() || !v.IsKnown() || v.Type() != cty.String {
		return ""
	}
	return v.AsString()
}

// walk follows a path of block names and list indices. It reports false when a
// step cannot be taken; an unknown value ends the walk successfully, so the
// caller can decide what unknown means.
func walk(v cty.Value, path []any) (cty.Value, bool) {
	for _, step := range path {
		if v.IsNull() {
			return cty.NilVal, false
		}
		if !v.IsKnown() {
			return v, true
		}
		switch s := step.(type) {
		case string:
			t := v.Type()
			if !t.IsObjectType() || !t.HasAttribute(s) {
				return cty.NilVal, false
			}
			v = v.GetAttr(s)
		case int:
			if !v.Type().IsListType() && !v.Type().IsTupleType() {
				return cty.NilVal, false
			}
			if v.LengthInt() <= s {
				return cty.NilVal, false
			}
			v = v.Index(cty.NumberIntVal(int64(s)))
		default:
			return cty.NilVal, false
		}
	}
	return v, true
}
