package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceKafkaAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_kafka_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "name", "test-kafka-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "description", "test kafka privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.resource_type", "Topic",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.resource_name", "*",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.pattern_type", "LITERAL",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.operation", "Read",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.permission_type", "ALLOW",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.host", "*",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.1.resource_type", "Group",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.1.pattern_type", "PREFIXED",
					),
				),
			},
			{
				Config: kafkaAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "name", "test-kafka-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "description", "updated kafka privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.operation", "All",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_privilege.test", "acls.0.permission_type", "DENY",
					),
				),
			},
		},
	})
}

func TestAccDataSourceKafkaAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessPrivilegeStep1() + kafkaAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_kafka_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_kafka_access_privilege.test", "name", "test-kafka-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_kafka_access_privilege.test", "acls.#", "2",
					),
				),
			},
		},
	})
}

func kafkaAccessPrivilegeStep1() string {
	return `
resource "hush_kafka_access_privilege" "test" {
  name        = "test-kafka-priv"
  description = "test kafka privilege"

  acls {
    resource_type   = "Topic"
    resource_name   = "*"
    pattern_type    = "LITERAL"
    operation       = "Read"
    permission_type = "ALLOW"
  }

  acls {
    resource_type   = "Group"
    resource_name   = "app-"
    pattern_type    = "PREFIXED"
    operation       = "Read"
    permission_type = "ALLOW"
    host            = "10.0.0.0/8"
  }
}
`
}

func kafkaAccessPrivilegeStep2() string {
	return `
resource "hush_kafka_access_privilege" "test" {
  name        = "test-kafka-priv-updated"
  description = "updated kafka privilege"

  acls {
    resource_type   = "Cluster"
    resource_name   = "kafka-cluster"
    pattern_type    = "LITERAL"
    operation       = "All"
    permission_type = "DENY"
  }
}
`
}

const kafkaAccessPrivilegeDataSource = `
data "hush_kafka_access_privilege" "test" {
  id = hush_kafka_access_privilege.test.id
}
`
