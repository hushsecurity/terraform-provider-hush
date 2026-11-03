package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const pinotPriv = "hush_pinot_access_privilege.test"

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Pinot; the routes are Elasticsearch's.
		ms.AliasRoutes("access_privileges/elasticsearch", "access_privileges/pinot")
		setType := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["type"] = "pinot"
			return nil
		}
		ms.OnOperation("access_privileges/pinot", testutil.OpCreate, setType)
		ms.OnOperation("access_privileges/pinot", testutil.OpUpdate, setType)
	})
}

func TestAccResourcePinotAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessPrivilege("test-pinot-priv", `tables = ["orders"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(pinotPriv, "id", regexp.MustCompile(`^apr-.+$`)),
					resource.TestCheckResourceAttr(pinotPriv, "name", "test-pinot-priv"),
					resource.TestCheckResourceAttr(pinotPriv, "type", "pinot"),
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.tables.#", "1"),
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.tables.0", "orders"),
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.exclude_tables.#", "0"),
				),
			},
			{
				Config: pinotAccessPrivilege("test-pinot-priv-updated", `exclude_tables = ["payroll"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pinotPriv, "name", "test-pinot-priv-updated"),
					// tables removed: all tables but the excluded one
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.tables.#", "0"),
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.exclude_tables.0", "payroll"),
				),
			},
			{
				// An empty grant is all tables, and the plan converges on it.
				Config: pinotAccessPrivilege("test-pinot-priv-updated", ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pinotPriv, "grant.#", "1"),
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.tables.#", "0"),
					resource.TestCheckResourceAttr(pinotPriv, "grant.0.exclude_tables.#", "0"),
				),
			},
			{
				ResourceName:      pinotPriv,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSourcePinotAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessPrivilege("test-pinot-priv", `
    tables         = ["orders", "events"]
    exclude_tables = ["payroll"]`) + `
data "hush_pinot_access_privilege" "test" {
  id = hush_pinot_access_privilege.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_pinot_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_pinot_access_privilege.test", "grant.0.tables.1", "events",
					),
					resource.TestCheckResourceAttr(
						"data.hush_pinot_access_privilege.test", "grant.0.exclude_tables.0", "payroll",
					),
				),
			},
		},
	})
}

// Every plan-time refusal: each branch of validateGrant and each element
// validator. None reaches the mock.
func TestAccResourcePinotAccessPrivilege_GrantValidation(t *testing.T) {
	cases := map[string]struct {
		grant string
		err   string
	}{
		"empty tables":         {`tables = []`, `grant.tables cannot be an empty list`},
		"empty exclude_tables": {`exclude_tables = []`, `grant.exclude_tables cannot be an empty list`},
		"suffixed table":       {`tables = ["orders_OFFLINE"]`, `must be named without its _OFFLINE suffix`},
		"suffixed exclusion":   {`exclude_tables = ["t_REALTIME"]`, `must be named without its _REALTIME suffix`},
		"empty table name":     {`tables = [""]`, `cannot contain an empty table name`},
		"empty excluded name":  {`exclude_tables = [""]`, `cannot contain an empty table name`},
		// the user is a broker USER; nothing else can be granted
		"components":  {`components = ["CONTROLLER"]`, `An argument named "components" is not expected here`},
		"permissions": {`permissions = ["READ"]`, `An argument named "permissions" is not expected here`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{
					{
						Config:      pinotAccessPrivilege("test-pinot-priv", c.grant),
						ExpectError: regexp.MustCompile(c.err),
					},
				},
			})
		})
	}
}

func pinotAccessPrivilege(name, grant string) string {
	return `
resource "hush_pinot_access_privilege" "test" {
  name        = "` + name + `"
  description = "test pinot privilege"

  grant {
    ` + grant + `
  }
}
`
}
