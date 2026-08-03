package provider

import (
	"slices"
	"strings"
	"testing"
)

// RequiredWith is an AND with no "either of these" form, so an attribute that
// names the plain half of a secret pair makes the write-only half unreachable:
// the plain half is demanded, and it conflicts with the write-only one. That
// inversion shipped on hush_redis_access_credential's elasticache engine, where
// access_key_id required secret_access_key and secret_access_key_wo was
// therefore unusable. Pair such attributes in CustomizeDiff, which can accept
// either variant.
func TestWriteOnlySecretPairing(t *testing.T) {
	for name, res := range New("test")().ResourcesMap {
		for attr, s := range res.Schema {
			if !s.WriteOnly {
				continue
			}
			plain := strings.TrimSuffix(attr, "_wo")
			for other, os := range res.Schema {
				if slices.Contains(os.RequiredWith, plain) {
					t.Errorf("%s: %q lists %q in RequiredWith, so %q can never be used; "+
						"enforce the pair in CustomizeDiff instead", name, other, plain, attr)
				}
			}
		}
	}
}
