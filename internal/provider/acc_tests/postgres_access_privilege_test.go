package acc_tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourcePostgresAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_postgres_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "name", "test-pg-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "description", "test postgres privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.object_type", "TABLE",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.privileges.0", "SELECT",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.privileges.1", "INSERT",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.all_in_schema", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.object_names.0", "public",
					),
				),
			},
			{
				Config: postgresAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_postgres_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "name", "test-pg-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "description", "updated postgres privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourcePostgresAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessPrivilegeStep1() + postgresAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_postgres_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_privilege.test", "name", "test-pg-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_privilege.test", "grants.0.object_type", "TABLE",
					),
				),
			},
		},
	})
}

func postgresAccessPrivilegeStep1() string {
	return `
resource "hush_postgres_access_privilege" "test" {
  name        = "test-pg-priv"
  description = "test postgres privilege"

  grants {
    privileges    = ["SELECT", "INSERT"]
    object_type   = "TABLE"
    all_in_schema = true
    object_names  = ["public"]
  }
}
`
}

func postgresAccessPrivilegeStep2() string {
	return `
resource "hush_postgres_access_privilege" "test" {
  name        = "test-pg-priv-updated"
  description = "updated postgres privilege"

  grants {
    privileges    = ["SELECT", "INSERT"]
    object_type   = "TABLE"
    all_in_schema = true
    object_names  = ["public"]
  }
}
`
}

const postgresAccessPrivilegeDataSource = `
data "hush_postgres_access_privilege" "test" {
  id = hush_postgres_access_privilege.test.id
}
`

// member_of stands beside grants or alone, and clearing it sends an empty
// list rather than leaving the stored roles in place.
func TestAccResourcePostgresAccessPrivilege_MemberOf(t *testing.T) {
	const name = "hush_postgres_access_privilege.member_of"
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessPrivilegeMemberOf(postgresSelectGrant, `["app_owner", "pg_monitor"]`),
				Check: resource.ComposeTestCheckFunc(
					recordID(name, &id),
					resource.TestCheckResourceAttr(name, "grants.#", "1"),
					resource.TestCheckResourceAttr(name, "member_of.#", "2"),
					resource.TestCheckTypeSetElemAttr(name, "member_of.*", "app_owner"),
					resource.TestCheckTypeSetElemAttr(name, "member_of.*", "pg_monitor"),
				),
			},
			{
				Config: postgresAccessPrivilegeMemberOf("", `["app_owner"]`) + postgresAccessPrivilegeMemberOfDataSource,
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged(name, &id),
					resource.TestCheckResourceAttr(name, "grants.#", "0"),
					resource.TestCheckResourceAttr(name, "member_of.#", "1"),
					resource.TestCheckTypeSetElemAttr(name, "member_of.*", "app_owner"),
					resource.TestCheckResourceAttr("data."+name, "member_of.#", "1"),
					resource.TestCheckTypeSetElemAttr("data."+name, "member_of.*", "app_owner"),
				),
			},
			{
				Config: postgresAccessPrivilegeMemberOf(postgresSelectGrant, `[]`),
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged(name, &id),
					resource.TestCheckResourceAttr(name, "grants.#", "1"),
					resource.TestCheckResourceAttr(name, "member_of.#", "0"),
				),
			},
		},
	})
}

// Plan-time checks, before any request reaches the mock: a privilege granting
// nothing, and role names midgard refuses.
func TestAccResourcePostgresAccessPrivilege_MemberOfValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      postgresAccessPrivilegeMemberOf("", `[]`),
				ExpectError: regexp.MustCompile(`at least one of grants or member_of is required`),
			},
			{
				Config:      postgresAccessPrivilegeMemberOf("", `["public"]`),
				ExpectError: regexp.MustCompile(`expected member_of\.\d+ to not be any of \[public\], got public`),
			},
			{
				Config:      postgresAccessPrivilegeMemberOf("", `["`+strings.Repeat("r", 64)+`"]`),
				ExpectError: regexp.MustCompile(`expected length of member_of\.\d+ to be in the range \(1 - 63\)`),
			},
		},
	})
}

// A role known only at apply passes the plan-time check, which cannot tell
// whether the privilege grants anything yet.
func TestAccResourcePostgresAccessPrivilege_MemberOfUnknownAtPlan(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessPrivilegeMemberOf(postgresSelectGrant, `["app_owner"]`) + `
resource "hush_postgres_access_privilege" "unknown" {
  name      = "test-pg-member-of-unknown"
  member_of = [hush_postgres_access_privilege.member_of.id]
}
`,
				Check: resource.TestCheckResourceAttr(
					"hush_postgres_access_privilege.unknown", "member_of.#", "1",
				),
			},
		},
	})
}

const postgresSelectGrant = `
  grants {
    privileges   = ["SELECT"]
    object_type  = "TABLE"
    object_names = ["weather"]
  }
`

func postgresAccessPrivilegeMemberOf(grants, memberOf string) string {
	return fmt.Sprintf(`
resource "hush_postgres_access_privilege" "member_of" {
  name      = "test-pg-member-of"
  member_of = %s
%s}
`, memberOf, grants)
}

const postgresAccessPrivilegeMemberOfDataSource = `
data "hush_postgres_access_privilege" "member_of" {
  id = hush_postgres_access_privilege.member_of.id
}
`
