package acc_tests

import (
	"fmt"
	"maps"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// statsigPatchedPrivileges holds, per privilege id, the object its last PATCH
// left in the mock.
var statsigPatchedPrivileges sync.Map

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Statsig; the routes are SendGrid's.
		ms.AliasRoutes("access_privileges/sendgrid", "access_privileges/statsig")
		setType := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["type"] = "statsig"
			return nil
		}
		ms.OnOperation("access_privileges/statsig", testutil.OpCreate, setType)
		ms.OnOperation("access_privileges/statsig", testutil.OpUpdate, setType)
		ms.OnOperation("access_privileges/statsig", testutil.OpUpdate,
			func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
				if id, ok := obj["id"].(string); ok {
					statsigPatchedPrivileges.Store(id, maps.Clone(obj))
				}
				return nil
			})
	})
}

const statsigPrivilege = "hush_statsig_access_privilege.test"

func TestAccResourceStatsigAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("statsig_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: statsigAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(statsigPrivilege, "id", regexp.MustCompile(`^apr-.+$`)),
					resource.TestCheckResourceAttr(statsigPrivilege, "name", "test-statsig-priv"),
					resource.TestCheckResourceAttr(statsigPrivilege, "key_type", "server"),
					resource.TestCheckResourceAttr(statsigPrivilege, "environments.#", "0"),
					resource.TestCheckResourceAttr(statsigPrivilege, "target_app_id", ""),
				),
			},
			{
				Config: statsigAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statsigPrivilege, "name", "test-statsig-priv-updated"),
					resource.TestCheckResourceAttr(statsigPrivilege, "key_type", "client"),
					resource.TestCheckResourceAttr(statsigPrivilege, "environments.#", "1"),
					resource.TestCheckTypeSetElemAttr(statsigPrivilege, "environments.*", "production"),
					resource.TestCheckResourceAttr(statsigPrivilege, "target_app_id", "app-1"),
					resource.TestCheckResourceAttr(statsigPrivilege, "secondary_target_app_ids.#", "1"),
				),
			},
			{
				// Removing the scoping widens the key again.
				Config: statsigAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					checkStatsigPatchedPrivilege(map[string]any{
						"key_type":                 "server",
						"environments":             nil,
						"target_app_id":            nil,
						"secondary_target_app_ids": nil,
					}),
					resource.TestCheckResourceAttr(statsigPrivilege, "key_type", "server"),
					resource.TestCheckResourceAttr(statsigPrivilege, "environments.#", "0"),
					resource.TestCheckResourceAttr(statsigPrivilege, "target_app_id", ""),
					resource.TestCheckResourceAttr(statsigPrivilege, "secondary_target_app_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceStatsigAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("statsig_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: statsigAccessPrivilegeStep2() + statsigAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data."+statsigPrivilege, "name", "test-statsig-priv-updated"),
					resource.TestCheckResourceAttr("data."+statsigPrivilege, "key_type", "client"),
					resource.TestCheckResourceAttr("data."+statsigPrivilege, "target_app_id", "app-1"),
				),
			},
		},
	})
}

func TestAccStatsigAccessPrivilegeRefusesBadInput(t *testing.T) {
	for _, tc := range []struct{ attr, want string }{
		{`key_type = "console"`, `expected key_type to be one of`},
		{`environments = []`, `Attribute environments requires 1 item minimum`},
		{`environments = [" production"]`, `must not start or end with whitespace`},
		{"environments = [\"production\u00a0\"]", `must not start or end with whitespace`},
		{`target_app_id = "app-1\u000b"`, `must not start or end with whitespace`},
		{`target_app_id = ""`, `expected "target_app_id" to not be an empty string`},
	} {
		t.Run(tc.attr, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config: `
resource "hush_statsig_access_privilege" "test" {
  name = "bad"
  ` + tc.attr + `
}
`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(regexp.QuoteMeta(tc.want)),
				}},
			})
		})
	}
}

func statsigAccessPrivilegeStep1() string {
	return `
resource "hush_statsig_access_privilege" "test" {
  name        = "test-statsig-priv"
  description = "test statsig privilege"
}
`
}

func statsigAccessPrivilegeStep2() string {
	return `
resource "hush_statsig_access_privilege" "test" {
  name                     = "test-statsig-priv-updated"
  description              = "updated statsig privilege"
  key_type                 = "client"
  environments             = ["production"]
  target_app_id            = "app-1"
  secondary_target_app_ids = ["app-2"]
}
`
}

const statsigAccessPrivilegeDataSource = `
data "hush_statsig_access_privilege" "test" {
  id = hush_statsig_access_privilege.test.id
}
`

// checkStatsigPatchedPrivilege checks the fields the last PATCH left in the
// mock, which the resource's state cannot show was sent.
func checkStatsigPatchedPrivilege(want map[string]any) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[statsigPrivilege]
		if !ok {
			return fmt.Errorf("resource not found: %s", statsigPrivilege)
		}
		v, ok := statsigPatchedPrivileges.Load(rs.Primary.ID)
		if !ok {
			return fmt.Errorf("no PATCH reached %s", rs.Primary.ID)
		}
		got := v.(map[string]any)
		for k, w := range want {
			if fmt.Sprint(got[k]) != fmt.Sprint(w) {
				return fmt.Errorf("PATCH left %s = %v, want %v", k, got[k], w)
			}
		}
		return nil
	}
}
