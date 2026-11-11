package acc_tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAgentAuthorizationServer(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "hush_agent_authorization_server" "test" {
  deployment_id = "dep-as-1"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hush_agent_authorization_server.test",
						"issuer", "https://api.example.com/v1/deployments/dep-as-1/agents"),
					resource.TestCheckResourceAttr("data.hush_agent_authorization_server.test",
						"token_endpoint", "https://api.example.com/v1/agents/oauth/token"),
				),
			},
		},
	})
}
