package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceElasticsearchAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_elasticsearch_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "name", "test-es-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "description", "test elasticsearch privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "grant.0.cluster.0", "monitor",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "grant.0.indices.0.names.0", "logs-*",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "grant.0.indices.0.privileges.0", "read",
					),
				),
			},
			{
				Config: elasticsearchAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_elasticsearch_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "name", "test-es-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "description", "updated elasticsearch privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.test", "grant.0.cluster.0", "manage",
					),
				),
			},
		},
	})
}

func TestAccResourceElasticsearchAccessPrivilege_indicesOnly(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessPrivilegeIndicesOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_elasticsearch_access_privilege.indices_only", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.indices_only", "grant.0.indices.0.names.0", "app-*",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_privilege.indices_only", "grant.0.indices.0.privileges.0", "all",
					),
				),
			},
		},
	})
}

func TestAccDataSourceElasticsearchAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessPrivilegeStep1() + elasticsearchAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_elasticsearch_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_elasticsearch_access_privilege.test", "name", "test-es-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_elasticsearch_access_privilege.test", "grant.0.cluster.0", "monitor",
					),
				),
			},
		},
	})
}

func elasticsearchAccessPrivilegeStep1() string {
	return `
resource "hush_elasticsearch_access_privilege" "test" {
  name        = "test-es-priv"
  description = "test elasticsearch privilege"

  grant {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*"]
      privileges = ["read"]
    }
  }
}
`
}

func elasticsearchAccessPrivilegeStep2() string {
	return `
resource "hush_elasticsearch_access_privilege" "test" {
  name        = "test-es-priv-updated"
  description = "updated elasticsearch privilege"

  grant {
    cluster = ["manage"]

    indices {
      names      = ["logs-*", "metrics-*"]
      privileges = ["read", "write"]
    }
  }
}
`
}

func elasticsearchAccessPrivilegeIndicesOnly() string {
	return `
resource "hush_elasticsearch_access_privilege" "indices_only" {
  name        = "test-es-indices-only"
  description = "indices only privilege"

  grant {
    indices {
      names      = ["app-*"]
      privileges = ["all"]
    }
  }
}
`
}

const elasticsearchAccessPrivilegeDataSource = `
data "hush_elasticsearch_access_privilege" "test" {
  id = hush_elasticsearch_access_privilege.test.id
}
`
