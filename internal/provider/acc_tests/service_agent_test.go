package acc_tests

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	agentAWSIssuer = "https://a1c0f603.tokens.sts.global.api.aws"
	agentTenant    = "8ea310af-fd38-43b1-8ff6-17df102dc19a"
	agentSP        = "1255db70-bd31-4355-86f6-fadfb99d9279"
	agentApp       = "609677e2-6afc-4fc1-a44b-0e5e5d662b3e"
)

// agentHolds checks the credential types heimdall holds for an agent.
func agentHolds(name string, want ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		slices.Sort(want)
		if got := agentsFake.credentialTypes(name); !slices.Equal(got, want) {
			return fmt.Errorf("%s holds %v, want %v", name, got, want)
		}
		return nil
	}
}

func agentsDestroyed(*terraform.State) error {
	agentsFake.mu.Lock()
	defer agentsFake.mu.Unlock()
	for id, agent := range agentsFake.agents {
		if agent["name"] == "acc-summarizer" || agent["name"] == "acc-summarizer-2" {
			return fmt.Errorf("agent %s still exists", id)
		}
	}
	return nil
}

const serviceAgentStep1 = `
resource "hush_service_agent" "test" {
  name = "acc-summarizer"

  aws_iam_role {
    issuer   = "` + agentAWSIssuer + `"
    role_arn = "arn:aws:iam::123456789012:role/summarizer"
  }

  azure {
    tenant_id            = "` + agentTenant + `"
    service_principal_id = "` + agentSP + `"
    audience             = "` + agentApp + `"

    conditions {
      claim = ["xms_mirid"]
      op    = "sfx"
      value = "/summarizer"
    }
  }
}

resource "hush_service_agent_client_secret" "test" {
  agent_id = hush_service_agent.test.id
}
`

// Renamed and acting for users; the AWS role goes, a Kubernetes service
// account and a plain OIDC identity come, and the Azure credential changes,
// which replaces it. The client secret is managed apart and stays.
const serviceAgentStep2 = `
resource "hush_service_agent" "test" {
  name           = "acc-summarizer-2"
  acts_for_users = true

  kubernetes {
    issuer          = "https://oidc.eks.eu-central-1.amazonaws.com/id/ABC"
    namespace       = "agents"
    service_account = "summarizer"
  }

  oidc {
    issuer  = "https://token.actions.githubusercontent.com"
    subject = "repo:acme/summarizer:ref:refs/heads/main"
  }

  azure {
    tenant_id            = "` + agentTenant + `"
    service_principal_id = "` + agentSP + `"
    audience             = "` + agentApp + `"
  }
}

resource "hush_service_agent_client_secret" "test" {
  agent_id = hush_service_agent.test.id
}

data "hush_service_agent" "by_name" {
  name = hush_service_agent.test.name
}

data "hush_service_agent" "by_id" {
  id = hush_service_agent.test.id
}
`

func TestAccServiceAgent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      agentsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: serviceAgentStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_service_agent.test", "enabled", "true"),
					resource.TestCheckResourceAttr("hush_service_agent.test", "acts_for_users", "false"),
					resource.TestMatchResourceAttr("hush_service_agent.test", "resource",
						regexp.MustCompile(`/v1/agents/agt-`)),
					resource.TestCheckResourceAttr("hush_service_agent.test", "aws_iam_role.#", "1"),
					resource.TestCheckResourceAttr("hush_service_agent.test", "azure.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("hush_service_agent.test", "azure.*",
						map[string]string{"audience": agentApp, "conditions.0.op": "sfx"}),
					resource.TestMatchResourceAttr("hush_service_agent_client_secret.test", "secret",
						regexp.MustCompile(`^hush_as_`)),
					resource.TestCheckResourceAttrSet("hush_service_agent_client_secret.test", "hint"),
					agentHolds("acc-summarizer", "aws_iam_role", "azure", "client_secret"),
				),
			},
			{
				Config: serviceAgentStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_service_agent.test", "name", "acc-summarizer-2"),
					resource.TestCheckResourceAttr("hush_service_agent.test", "acts_for_users", "true"),
					resource.TestCheckResourceAttr("hush_service_agent.test", "aws_iam_role.#", "0"),
					resource.TestCheckResourceAttr("hush_service_agent.test", "kubernetes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("hush_service_agent.test", "azure.*",
						map[string]string{"conditions.#": "0"}),
					resource.TestMatchResourceAttr("hush_service_agent_client_secret.test", "secret",
						regexp.MustCompile(`^hush_as_`)),
					resource.TestCheckTypeSetElemNestedAttrs("hush_service_agent.test", "oidc.*",
						map[string]string{"issuer": "https://token.actions.githubusercontent.com"}),
					agentHolds("acc-summarizer-2", "azure", "client_secret", "kubernetes", "oidc"),
					resource.TestCheckResourceAttrPair("data.hush_service_agent.by_name", "id",
						"hush_service_agent.test", "id"),
					resource.TestCheckResourceAttr("data.hush_service_agent.by_name", "kubernetes.#", "1"),
					resource.TestCheckResourceAttr("data.hush_service_agent.by_id", "name", "acc-summarizer-2"),
					resource.TestCheckResourceAttr("data.hush_service_agent.by_id", "oidc.#", "1"),
				),
			},
			{
				ResourceName:      "hush_service_agent.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// The secret comes back only when it is created.
				ResourceName:            "hush_service_agent_client_secret.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccServiceAgentRefusesBadInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_service_agent" "bad" {
  name = "acc-bad"
  azure {
    tenant_id            = "8EA310AF-FD38-43B1-8FF6-17DF102DC19A"
    service_principal_id = "` + agentSP + `"
    audience             = "api://hush"
  }
}`,
				ExpectError: regexp.MustCompile(`must be a lowercase GUID`),
			},
			{
				Config: `
resource "hush_service_agent" "bad" {
  name = "acc-bad"
  aws_iam_role {
    issuer   = "http://insecure.example.com"
    role_arn = "arn:aws:iam::123456789012:user/alice"
  }
}`,
				ExpectError: regexp.MustCompile(`must be an IAM role ARN`),
			},
		},
	})
}

func TestAccServiceAgentRefusesBadPlans(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_service_agent" "bad" {
  name = "acc-bad"
  kubernetes {
    issuer          = "https://oidc.example.com"
    namespace       = "agents"
    service_account = "summarizer"
  }
  kubernetes {
    issuer          = "https://oidc.example.com"
    namespace       = "agents"
    service_account = "summarizer"
    conditions {
      claim = ["aud"]
      value = "hush"
    }
  }
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`the same identity is written twice`),
			},
			{
				// heimdall binds the issuer and subject, whatever the block.
				Config: `
resource "hush_service_agent" "bad" {
  name = "acc-bad"
  kubernetes {
    issuer          = "https://oidc.example.com"
    namespace       = "agents"
    service_account = "summarizer"
  }
  oidc {
    issuer  = "https://oidc.example.com"
    subject = "system:serviceaccount:agents:summarizer"
  }
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`the same identity is written twice`),
			},
			{
				Config:      oidcAgent("acc-bad", 0, 11),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`at most 10 identities, got 11`),
			},
		},
	})
}

// oidcAgent writes an agent with OIDC identities for subjects from..to-1.
func oidcAgent(name string, from, to int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "resource \"hush_service_agent\" \"test\" {\n  name = %q\n", name)
	for i := from; i < to; i++ {
		fmt.Fprintf(&b, "  oidc {\n    issuer  = \"https://oidc.example.com\"\n    subject = \"workload-%d\"\n  }\n", i)
	}
	b.WriteString("}\n")
	return b.String()
}

// At heimdall's limit, a changed identity can't be added before the old one
// goes, so the old one makes room.
func TestAccServiceAgentSwapsAtLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      agentsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: oidcAgent("acc-summarizer", 0, 10),
				Check:  resource.TestCheckResourceAttr("hush_service_agent.test", "oidc.#", "10"),
			},
			{
				Config: oidcAgent("acc-summarizer", 1, 11),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_service_agent.test", "oidc.#", "10"),
					resource.TestCheckTypeSetElemNestedAttrs("hush_service_agent.test", "oidc.*",
						map[string]string{"subject": "workload-10"}),
				),
			},
		},
	})
}

// A user's agent shares the agents API; it is never read as a service agent.
func TestAccServiceAgentRefusesUserAgent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      `data "hush_service_agent" "user" { id = "` + userAgentID + `" }`,
				ExpectError: regexp.MustCompile(`not a service agent`),
			},
			{
				Config:      `data "hush_service_agent" "user" { name = "alice@example.com" }`,
				ExpectError: regexp.MustCompile(`no service agent found with name`),
			},
			{
				Config:        `resource "hush_service_agent" "user" { name = "alice@example.com" }`,
				ResourceName:  "hush_service_agent.user",
				ImportState:   true,
				ImportStateId: userAgentID,
				ExpectError:   regexp.MustCompile(`not a service agent`),
			},
		},
	})
}

// Roles made in the same apply are unknown at plan time; two of them are not
// the same identity.
func TestAccServiceAgentPlansUnknownIdentities(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{"terraform": {Source: "terraform.io/builtin/terraform"}},
		CheckDestroy:      agentsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "a" {
  input = "arn:aws:iam::123456789012:role/summarizer-a"
}

resource "terraform_data" "b" {
  input = "arn:aws:iam::123456789012:role/summarizer-b"
}

resource "hush_service_agent" "test" {
  name = "acc-summarizer"
  aws_iam_role {
    issuer   = "` + agentAWSIssuer + `"
    role_arn = terraform_data.a.output
  }
  aws_iam_role {
    issuer   = "` + agentAWSIssuer + `"
    role_arn = terraform_data.b.output
  }
}`,
				Check: agentHolds("acc-summarizer", "aws_iam_role", "aws_iam_role"),
			},
		},
	})
}

func oidcAgents(subjectA, subjectB string) string {
	agent := func(name, subject string) string {
		return fmt.Sprintf(`
resource "hush_service_agent" %[1]q {
  name = "acc-summarizer-%[1]s"
  oidc {
    issuer  = "https://oidc.example.com"
    subject = %[2]q
  }
}
`, name, subject)
	}
	return agent("a", subjectA) + agent("b", subjectB)
}

// A refused add leaves the agent's old identity in place, and moving an
// identity between agents is refused until the other lets it go.
func TestAccServiceAgentKeepsIdentityWhenAddFails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{Config: oidcAgents("s1", "s2")},
			{
				Config:      oidcAgents("s2", "s2"),
				ExpectError: regexp.MustCompile(`identity already bound`),
			},
			{Config: oidcAgents("s1", "s2"), PlanOnly: true},
			// A kubernetes block may take over the same identity from an oidc one.
			{Config: oidcAgents("system:serviceaccount:agents:a", "s2")},
			{
				Config: strings.Replace(oidcAgents("system:serviceaccount:agents:a", "s2"), `oidc {
    issuer  = "https://oidc.example.com"
    subject = "system:serviceaccount:agents:a"
  }`, `kubernetes {
    issuer          = "https://oidc.example.com"
    namespace       = "agents"
    service_account = "a"
  }`, 1),
				Check: resource.TestCheckResourceAttr("hush_service_agent.a", "kubernetes.#", "1"),
			},
		},
	})
}
