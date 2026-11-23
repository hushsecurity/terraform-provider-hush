package acc_tests

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const r2Priv = "hush_cloudflare_r2_access_privilege.test"

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Cloudflare R2; the routes are SendGrid's.
		ms.AliasRoutes("access_privileges/sendgrid", "access_privileges/cloudflare_r2")
		ms.OnOperation("access_privileges/cloudflare_r2", testutil.OpCreate, r2PrivilegeHook(true))
		ms.OnOperation("access_privileges/cloudflare_r2", testutil.OpUpdate, r2PrivilegeHook(false))
	})
}

// r2PrivilegeHook copies the Hush API: it sorts the buckets, and refuses a list
// on an admin permission and, on create, an object permission without one.
func r2PrivilegeHook(create bool) testutil.HookFunc {
	return func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
		obj["type"] = "cloudflare_r2"
		permission, _ := obj["permission"].(string)
		admin := strings.HasPrefix(permission, "admin_")
		buckets, sent := obj["buckets"]
		if create && !admin && !sent {
			return &testutil.HookError{Status: http.StatusUnprocessableEntity,
				Detail: "buckets is required with an object permission"}
		}
		list, isList := buckets.([]any)
		if !isList {
			return nil
		}
		if admin {
			return &testutil.HookError{Status: http.StatusUnprocessableEntity,
				Detail: permission + " is account wide and takes no buckets; set buckets to null"}
		}
		sort.Slice(list, func(i, j int) bool { return list[i].(string) < list[j].(string) })
		return nil
	}
}

func TestAccResourceCloudflareR2AccessPrivilege(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				// the API sorts the buckets; the set shows no change for it
				Config: cloudflareR2AccessPrivilege("test-r2-priv", "object_read_write",
					`buckets = ["photos", "exports"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(r2Priv, "id", regexp.MustCompile(`^apr-.+$`)),
					resource.TestCheckResourceAttr(r2Priv, "name", "test-r2-priv"),
					resource.TestCheckResourceAttr(r2Priv, "type", "cloudflare_r2"),
					resource.TestCheckResourceAttr(r2Priv, "permission", "object_read_write"),
					resource.TestCheckResourceAttr(r2Priv, "buckets.#", "2"),
					resource.TestCheckTypeSetElemAttr(r2Priv, "buckets.*", "photos"),
					resource.TestCheckTypeSetElemAttr(r2Priv, "buckets.*", "exports"),
					resource.TestCheckResourceAttr(r2Priv, "all_buckets", "false"),
					resource.TestCheckResourceAttr(r2Priv, "jurisdiction", "default"),
					recordID(r2Priv, &id),
				),
			},
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv", "object_read_write",
					`buckets = ["logs", "photos", "exports"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Priv, "buckets.#", "3"),
					resource.TestCheckTypeSetElemAttr(r2Priv, "buckets.*", "logs"),
					checkIDUnchanged(r2Priv, &id),
				),
			},
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv", "object_read_write",
					`all_buckets = true`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Priv, "buckets.#", "0"),
					resource.TestCheckResourceAttr(r2Priv, "all_buckets", "true"),
					checkIDUnchanged(r2Priv, &id),
				),
			},
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv", "admin_read_only", ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Priv, "permission", "admin_read_only"),
					resource.TestCheckResourceAttr(r2Priv, "buckets.#", "0"),
					resource.TestCheckResourceAttr(r2Priv, "all_buckets", "false"),
					checkIDUnchanged(r2Priv, &id),
				),
			},
			{
				// from admin to object, the buckets go in the same request
				Config: cloudflareR2AccessPrivilege("test-r2-priv", "object_read_only",
					`buckets = ["photos"]`+"\n"+`jurisdiction = "eu"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Priv, "permission", "object_read_only"),
					resource.TestCheckResourceAttr(r2Priv, "buckets.#", "1"),
					resource.TestCheckResourceAttr(r2Priv, "all_buckets", "false"),
					resource.TestCheckResourceAttr(r2Priv, "jurisdiction", "eu"),
					checkIDUnchanged(r2Priv, &id),
				),
			},
			{
				// from listed buckets to admin, null clears them
				Config: cloudflareR2AccessPrivilege("test-r2-priv", "admin_read_write",
					`jurisdiction = "eu"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Priv, "permission", "admin_read_write"),
					resource.TestCheckResourceAttr(r2Priv, "buckets.#", "0"),
					checkIDUnchanged(r2Priv, &id),
				),
			},
			{
				ResourceName:      r2Priv,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceCloudflareR2AccessPrivilege_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv-import", "object_read_only",
					`all_buckets = true`),
			},
			{
				// all_buckets is read back from null buckets
				ResourceName:      r2Priv,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceCloudflareR2AccessPrivilege_Validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// every bucket is never granted by leaving buckets out
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_write", ``),
				ExpectError: regexp.MustCompile(`object_read_write needs buckets, or all_buckets = true`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_only", `all_buckets = false`),
				ExpectError: regexp.MustCompile(`object_read_only needs buckets, or all_buckets = true`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_write", `buckets = []`),
				ExpectError: regexp.MustCompile(`buckets cannot be empty`),
			},
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_write",
					`buckets = ["photos"]`+"\n"+`all_buckets = true`),
				ExpectError: regexp.MustCompile(`set buckets or all_buckets = true, not both`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "admin_read_write", `buckets = ["photos"]`),
				ExpectError: regexp.MustCompile(`admin_read_write is account wide and takes no buckets`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "admin_read_only", `all_buckets = true`),
				ExpectError: regexp.MustCompile(`admin_read_only is account wide and takes no buckets`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_write", `buckets = ["Photos"]`),
				ExpectError: regexp.MustCompile(`a bucket name must be 3-63 lowercase letters`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_write", `buckets = ["photos*"]`),
				ExpectError: regexp.MustCompile(`a bucket name must be 3-63 lowercase letters`),
			},
			{
				Config:      cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_write", `all_buckets = true`),
				ExpectError: regexp.MustCompile(`expected permission to be one of`),
			},
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv-bad", "object_read_write",
					`all_buckets = true`+"\n"+`jurisdiction = "apac"`),
				ExpectError: regexp.MustCompile(`expected jurisdiction to be one of`),
			},
			{
				// an explicit all_buckets = false with buckets passes
				Config: cloudflareR2AccessPrivilege("test-r2-priv-ok", "object_read_write",
					`buckets = ["photos"]`+"\n"+`all_buckets = false`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSourceCloudflareR2AccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessPrivilege("test-r2-priv-ds", "object_read_only",
					`buckets = ["photos", "exports"]`+"\n"+`jurisdiction = "fedramp"`) + `
data "hush_cloudflare_r2_access_privilege" "test" {
  id = hush_cloudflare_r2_access_privilege.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.hush_cloudflare_r2_access_privilege.test", "id", r2Priv, "id",
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_privilege.test", "permission", "object_read_only",
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_privilege.test", "buckets.#", "2",
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_privilege.test", "all_buckets", "false",
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_privilege.test", "jurisdiction", "fedramp",
					),
				),
			},
		},
	})
}

// cloudflareR2AccessPrivilege renders a privilege with the given permission
// and grant attributes.
func cloudflareR2AccessPrivilege(name, permission, attrs string) string {
	return `
resource "hush_cloudflare_r2_access_privilege" "test" {
  name       = "` + name + `"
  permission = "` + permission + `"
  ` + attrs + `
}
`
}
