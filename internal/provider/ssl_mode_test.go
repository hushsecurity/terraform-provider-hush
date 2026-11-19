package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/mariadb_access_credential"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/mysql_access_credential"
)

// Both resources take ssl_mode in the API's hyphenated form; mysql took
// verify_ca and verify_identity, which the API refuses at apply.
func TestSSLModeValidatorTakesHyphenatedForms(t *testing.T) {
	for name, s := range map[string]map[string]*schema.Schema{
		"mysql":   mysql_access_credential.ResourceSchema(),
		"mariadb": mariadb_access_credential.ResourceSchema(),
	} {
		validate := s["ssl_mode"].ValidateFunc
		for mode, ok := range map[string]bool{
			"disabled":        true,
			"preferred":       true,
			"required":        true,
			"verify-ca":       true,
			"verify-identity": true,
			"verify_ca":       false,
			"verify_identity": false,
		} {
			_, errs := validate(mode, "ssl_mode")
			if (len(errs) == 0) != ok {
				t.Errorf("%s ssl_mode %q: want accepted=%v, got %v", name, mode, ok, errs)
			}
		}
	}
}
