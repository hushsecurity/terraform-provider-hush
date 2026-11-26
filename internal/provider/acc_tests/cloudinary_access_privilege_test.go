package acc_tests

import (
	"net/http"
	"regexp"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const (
	cldPriv   = "hush_cloudinary_access_privilege.test"
	cldEditor = "cld::role::prodenv::editor"
	cldUpload = "cld::policy::global::folder_and_asset_management::create_asset"
	cldPreset = "cld::policy::global::upload_presets::manage"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Cloudinary; the routes are SendGrid's.
		ms.AliasRoutes("access_privileges/sendgrid", "access_privileges/cloudinary")
		ms.OnOperation("access_privileges/cloudinary", testutil.OpCreate, cloudinaryPrivilegeHook)
		ms.OnOperation("access_privileges/cloudinary", testutil.OpUpdate, cloudinaryPrivilegeHook)
	})
}

// cloudinaryPrivilegeHook copies the Hush API: it sorts the policies and refuses
// both or neither grant, so a switch must send the old grant as null.
func cloudinaryPrivilegeHook(_ testutil.Operation, obj map[string]any) *testutil.HookError {
	obj["type"] = "cloudinary"
	role, hasRole := obj["role_id"].(string)
	policies, hasPolicies := obj["system_policy_ids"].([]any)
	if (hasRole && role != "") == hasPolicies {
		return &testutil.HookError{Status: http.StatusUnprocessableEntity,
			Detail: "set exactly one of role_id and system_policy_ids"}
	}
	sort.Slice(policies, func(i, j int) bool { return policies[i].(string) < policies[j].(string) })
	return nil
}

func TestAccResourceCloudinaryAccessPrivilege(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessPrivilege("test-cld-priv", "product1", `role_id = "`+cldEditor+`"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(cldPriv, "id", regexp.MustCompile(`^apr-.+$`)),
					resource.TestCheckResourceAttr(cldPriv, "name", "test-cld-priv"),
					resource.TestCheckResourceAttr(cldPriv, "type", "cloudinary"),
					resource.TestCheckResourceAttr(cldPriv, "cloud_name", "product1"),
					resource.TestCheckResourceAttr(cldPriv, "role_id", cldEditor),
					resource.TestCheckResourceAttr(cldPriv, "system_policy_ids.#", "0"),
					recordID(cldPriv, &id),
				),
			},
			{
				// from the role to policies, the role goes as null in the
				// same request; the API sorts the policies, which the set ignores
				Config: cloudinaryAccessPrivilege("test-cld-priv", "product1",
					`system_policy_ids = ["`+cldUpload+`", "`+cldPreset+`"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldPriv, "role_id", ""),
					resource.TestCheckResourceAttr(cldPriv, "system_policy_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(cldPriv, "system_policy_ids.*", cldUpload),
					resource.TestCheckTypeSetElemAttr(cldPriv, "system_policy_ids.*", cldPreset),
					checkIDUnchanged(cldPriv, &id),
				),
			},
			{
				// another order is no change
				Config: cloudinaryAccessPrivilege("test-cld-priv", "product1",
					`system_policy_ids = ["`+cldPreset+`", "`+cldUpload+`"]`),
				PlanOnly: true,
			},
			{
				Config: cloudinaryAccessPrivilege("test-cld-priv", "product2",
					`system_policy_ids = ["`+cldUpload+`"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldPriv, "cloud_name", "product2"),
					resource.TestCheckResourceAttr(cldPriv, "system_policy_ids.#", "1"),
					checkIDUnchanged(cldPriv, &id),
				),
			},
			{
				// back to a role, the policies go as null
				Config: cloudinaryAccessPrivilege("test-cld-priv", "product2",
					`role_id = "cld::role::prodenv::viewer"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldPriv, "role_id", "cld::role::prodenv::viewer"),
					resource.TestCheckResourceAttr(cldPriv, "system_policy_ids.#", "0"),
					checkIDUnchanged(cldPriv, &id),
				),
			},
			{
				ResourceName:      cldPriv,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceCloudinaryAccessPrivilege_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessPrivilege("test-cld-priv-import", "product1",
					`system_policy_ids = ["`+cldUpload+`", "`+cldPreset+`"]`),
			},
			{
				ResourceName:      cldPriv,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceCloudinaryAccessPrivilege_Validation(t *testing.T) {
	const name = "test-cld-priv-bad"
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      cloudinaryAccessPrivilege(name, "product1", ``),
				ExpectError: regexp.MustCompile(`set exactly one of role_id and system_policy_ids`),
			},
			{
				Config: cloudinaryAccessPrivilege(name, "product1",
					`role_id = "`+cldEditor+`"`+"\n"+`system_policy_ids = ["`+cldUpload+`"]`),
				ExpectError: regexp.MustCompile(`set exactly one of role_id and system_policy_ids`),
			},
			{
				Config:      cloudinaryAccessPrivilege(name, "product1", `system_policy_ids = []`),
				ExpectError: regexp.MustCompile(`system_policy_ids cannot be empty`),
			},
			{
				// content policies need a folder or collection
				Config: cloudinaryAccessPrivilege(name, "product1",
					`system_policy_ids = ["cld::policy::content::folder::view"]`),
				ExpectError: regexp.MustCompile(`is not a cld::policy::global:: policy`),
			},
			{
				Config:      cloudinaryAccessPrivilege(name, "1product", `role_id = "`+cldEditor+`"`),
				ExpectError: regexp.MustCompile(`cloud_name must be 2 to 128 letters, digits or '-'`),
			},
		},
	})
}

// The data source reads back what the API stored, so it shows an update was sent.
func TestAccDataSourceCloudinaryAccessPrivilege(t *testing.T) {
	const data = "data.hush_cloudinary_access_privilege.test"
	dataSource := `
data "hush_cloudinary_access_privilege" "test" {
  id = hush_cloudinary_access_privilege.test.id
}
`
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessPrivilege("test-cld-priv-ds", "product1",
					`role_id = "`+cldEditor+`"`) + dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(data, "id", cldPriv, "id"),
					resource.TestCheckResourceAttr(data, "cloud_name", "product1"),
					resource.TestCheckResourceAttr(data, "role_id", cldEditor),
					resource.TestCheckResourceAttr(data, "system_policy_ids.#", "0"),
				),
			},
			{
				Config: cloudinaryAccessPrivilege("test-cld-priv-ds", "product2",
					`system_policy_ids = ["`+cldUpload+`", "`+cldPreset+`"]`) + dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data, "cloud_name", "product2"),
					resource.TestCheckResourceAttr(data, "role_id", ""),
					resource.TestCheckResourceAttr(data, "system_policy_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(data, "system_policy_ids.*", cldPreset),
				),
			},
		},
	})
}

// cloudinaryAccessPrivilege renders a privilege with the given cloud name and
// grant attributes.
func cloudinaryAccessPrivilege(name, cloudName, attrs string) string {
	return `
resource "hush_cloudinary_access_privilege" "test" {
  name       = "` + name + `"
  cloud_name = "` + cloudName + `"
  ` + attrs + `
}
`
}
