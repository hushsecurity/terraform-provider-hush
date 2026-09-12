package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// Mirror the API's rules on the agw object. The mock accepts anything on its
// own, so without these hooks the provider could send an empty object or a
// cleared hostname and every step below would still pass.
//
// Registered from init() so they cover every deployment write in the package,
// not only the steps of this test.
func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		refuseEmptyGateway := func(
			op testutil.Operation, obj map[string]any,
		) *testutil.HookError {
			agw, present := obj["agw"]
			if !present || agw == nil {
				return nil
			}
			fields, ok := agw.(map[string]any)
			if !ok {
				return nil
			}
			// API rule: an on-prem gateway is identified by its hostname, and
			// a record without one is refused on create and cannot be reached
			// by clearing it on update. A region instead of a hostname is a
			// Hush-hosted gateway, which this provider never writes.
			if region, ok := fields["region"].(string); ok && region != "" {
				return nil
			}
			host, ok := fields["hostname"].(string)
			if !ok || host == "" {
				return &testutil.HookError{
					Status: 422,
					Detail: "agw.hostname is required for an on-prem gateway",
				}
			}
			// The API derives the address from the hostname, so a move
			// changes it. Without this the mock leaves gateway_url empty
			// through every step and the tests could not see it go stale.
			fields["gateway_url"] = "https://" + host + "/mcp"
			return nil
		}

		ms.OnOperation("deployment", testutil.OpCreate, refuseEmptyGateway)
		ms.OnOperation("deployment", testutil.OpUpdate, refuseEmptyGateway)

		// A hosted gateway is placed by Hush, so the API derives the address
		// and the OIDC issuer of the cluster it runs in. Without this the mock
		// returns neither and the tests below could not tell a working read
		// path from one that drops both.
		ms.OnOperation("deployment", testutil.OpCreate,
			func(op testutil.Operation, obj map[string]any) *testutil.HookError {
				if obj["kind"] != deploymentKindHosted {
					return nil
				}
				agw, _ := obj["agw"].(map[string]any)
				if agw == nil {
					return &testutil.HookError{
						Status: 422,
						Detail: "agw is required for a hosted deployment",
					}
				}
				region, _ := agw["region"].(string)
				if region == "" {
					return &testutil.HookError{
						Status: 422,
						Detail: "agw.region is required for a hosted deployment",
					}
				}
				agw["gateway_url"] = fmt.Sprintf(
					"https://%s.agw.%s.example.com/mcp", obj["id"], region)
				obj["oidc_provider"] = map[string]any{
					"issuer":   "https://oidc.eks.eu-west-1.amazonaws.com/id/MOCK",
					"audience": "https://kubernetes.default.svc",
				}
				return nil
			})

		// The API rules a patch on a hosted deployment breaks. Each of these
		// is reachable only if the provider sends something it should not, so
		// they exist to make that loud rather than to be hit.
		ms.OnOperation("deployment", testutil.OpUpdate,
			func(op testutil.Operation, obj map[string]any) *testutil.HookError {
				if obj["kind"] != deploymentKindHosted {
					return nil
				}
				if obj["oidc_providers"] != nil {
					return &testutil.HookError{
						Status: 422,
						Detail: "oidc_providers is not valid for a hosted deployment",
					}
				}
				agw, _ := obj["agw"].(map[string]any)
				if agw == nil {
					return &testutil.HookError{
						Status: 422,
						Detail: "agw cannot be removed from a hosted deployment",
					}
				}
				if host, ok := agw["hostname"]; ok && host != nil {
					return &testutil.HookError{
						Status: 422,
						Detail: "agw.hostname is not valid for a hosted deployment",
					}
				}
				return nil
			})
	})
}

// The hosted tests key off the same kind string the provider uses.
const deploymentKindHosted = "hosted"

// TestAccResourceDeploymentAgw walks a customer-run gateway through its life:
// installed with the chart, moved to another hostname, decommissioned, then
// installed again on a deployment that already exists.
//
// Every step differs only in the agw block, so a failure cannot be blamed on
// anything else changing at the same time.
func TestAccResourceDeploymentAgw(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentAgwConfig(agwDeploymentName, "k8s", agwHostname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.0.hostname", agwHostname),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.0.gateway_url", "https://"+agwHostname+"/mcp"),
				),
			},
			{
				// DNS moves. The API allows the hostname to change, so this
				// must not force a new deployment -- replacing it would throw
				// away the token the gateway authenticates with.
				//
				// gateway_url is derived from the hostname, and the plan
				// carries no entry for it, so this is where it goes stale if
				// the update does not take its state from the response.
				Config: deploymentAgwConfig(agwDeploymentName, "k8s", agwHostname2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.0.hostname", agwHostname2),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.0.gateway_url", "https://"+agwHostname2+"/mcp"),
				),
			},
			{
				// The gateway is decommissioned and the sensor stays. Dropping
				// the block has to send an explicit null: an omitted field
				// leaves the stored object in place, and the deployment would
				// keep answering with a gateway that is gone.
				Config: deploymentAgwConfig(agwDeploymentName, "k8s"),
				Check: resource.TestCheckResourceAttr("hush_deployment.test",
					"agw.#", "0"),
			},
			{
				// A co-installed gateway joins a deployment created for the
				// sensor alone, which is the other way gateway_url is filled
				// in by an update rather than by a create.
				Config: deploymentAgwConfig(agwDeploymentName, "k8s", agwHostname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.0.hostname", agwHostname),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"agw.0.gateway_url", "https://"+agwHostname+"/mcp"),
				),
			},
		},
	})
}

// TestAccResourceDeploymentAgwRefusals covers what is refused before any
// request is sent, so a mistake in the hostname or the kind costs a plan rather
// than a round trip.
func TestAccResourceDeploymentAgwRefusals(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				// The gateway ships as a Helm chart, so no other self-hosted
				// kind can install it.
				Config:      deploymentAgwConfig(agwRefusalsDeploymentName, "ecs", agwHostname),
				ExpectError: regexp.MustCompile(`agw: a gateway is only valid on kind "k8s"`),
			},
			{
				// The API rejects mixed case rather than normalizing it, so
				// the provider must not accept what it cannot read back.
				Config:      deploymentAgwConfig(agwRefusalsDeploymentName, "k8s", "GW.example.com"),
				ExpectError: regexp.MustCompile(`must be lowercase`),
			},
			{
				// The chart asks for the same bare hostname, not a URL.
				Config:      deploymentAgwConfig(agwRefusalsDeploymentName, "k8s", "https://gw.example.com/mcp"),
				ExpectError: regexp.MustCompile(`must be a bare domain name`),
			},
		},
	})
}

// TestAccDataSourceDeploymentAgw verifies the data source surfaces the gateway,
// which is how a configuration elsewhere reaches a deployment it does not own.
func TestAccDataSourceDeploymentAgw(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentAgwConfig(agwDataDeploymentName, "k8s", agwHostname) +
					deploymentDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hush_deployment.test",
						"agw.#", "1"),
					resource.TestCheckResourceAttr("data.hush_deployment.test",
						"agw.0.hostname", agwHostname),
				),
			},
		},
	})
}

const (
	agwHostname  = "gw.example.com"
	agwHostname2 = "mcp-gateway.eu.example.com"

	agwDeploymentName         = "tf-acc-agw"
	agwRefusalsDeploymentName = "tf-acc-agw-refusals"
	agwDataDeploymentName     = "tf-acc-agw-data"
)

// deploymentAgwConfig renders a deployment at a fixed name, with an agw block
// only when a hostname is given, so no step of one test changes anything else.
// A name per test keeps the parallel tests off each other's deployments.
func deploymentAgwConfig(name, kind string, hostname ...string) string {
	block := ""
	if len(hostname) > 0 {
		block = fmt.Sprintf(`
  agw {
    hostname = %q
  }
`, hostname[0])
	}

	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = %q
  kind = %q
%s}
`, name, kind, block)
}

// TestAccResourceDeploymentHostedAgw covers a gateway Hush runs: the region
// places it, the address and the OIDC issuer come back derived, and neither is
// something a configuration has to state.
func TestAccResourceDeploymentHostedAgw(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentHostedAgwConfig(agwRegion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.hosted",
						"agw.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.hosted",
						"agw.0.region", agwRegion),
					resource.TestCheckResourceAttr("hush_deployment.hosted",
						"agw.0.hostname", ""),
					resource.TestMatchResourceAttr("hush_deployment.hosted",
						"agw.0.gateway_url", regexp.MustCompile(
							`^https://dep-.+\.agw\.`+agwRegion+`\.example\.com/mcp$`)),
					// The API sets the issuer itself and then refuses to be
					// sent either OIDC field. Surfacing it would put a block in
					// state that no configuration wrote, and the plan to remove
					// it could never apply -- the step's own no-empty-plan
					// check is what catches that.
					resource.TestCheckResourceAttr("hush_deployment.hosted",
						"oidc_provider.#", "0"),
				),
			},
			{
				// Placement is set once and the API has no field to change it.
				// Terraform refuses the edit rather than answering it with a
				// replacement, which would detach every application bound to
				// the gateway along with the old deployment id.
				Config: deploymentHostedAgwConfig(agwRegion2),
				ExpectError: regexp.MustCompile(
					`agw.region: cannot be changed from "` + agwRegion +
						`" to "` + agwRegion2 + `"`),
			},
			{
				// Restoring the region settles the plan, so the refusal costs
				// an edit rather than leaving the resource unmanageable.
				Config: deploymentHostedAgwConfig(agwRegion),
				Check: resource.TestCheckResourceAttr("hush_deployment.hosted",
					"agw.0.region", agwRegion),
			},
		},
	})
}

// TestAccResourceDeploymentAgwKindRules covers the pairing between the block
// and the kind. Each is refused before any request, so a configuration that
// describes the wrong sort of gateway costs a plan rather than a round trip.
func TestAccResourceDeploymentAgwKindRules(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				// Hush derives a hosted gateway's address.
				Config: deploymentAgwBlockConfig("tf-acc-agw-hosted-host",
					deploymentKindHosted,
					`hostname = "gw.example.com"`+"\n    "+`region = "`+agwRegion+`"`),
				ExpectError: regexp.MustCompile(
					`agw.hostname: not valid on kind "hosted"`),
			},
			{
				// Hush places nothing for a gateway the customer runs.
				Config: deploymentAgwBlockConfig("tf-acc-agw-k8s-region", "k8s",
					`hostname = "gw.example.com"`+"\n    "+`region = "`+agwRegion+`"`),
				ExpectError: regexp.MustCompile(
					`agw.region: not valid on kind "k8s"`),
			},
			{
				// A hosted gateway has to be placed somewhere. An empty block
				// rather than an empty hostname, which is refused earlier as a
				// malformed name.
				Config: deploymentAgwBlockConfig("tf-acc-agw-hosted-noregion",
					deploymentKindHosted, ""),
				ExpectError: regexp.MustCompile(
					`agw.region: required on kind "hosted"`),
			},
			{
				// The kind is what asks for a gateway Hush runs, so the block
				// cannot be left out.
				Config: deploymentNoAgwConfig("tf-acc-agw-hosted-noblock",
					deploymentKindHosted),
				ExpectError: regexp.MustCompile(
					`agw: an agw block with a region is required on kind "hosted"`),
			},
			{
				// The cluster a hosted gateway runs in is what it trusts, so
				// the caller does not get to name an issuer.
				Config: deploymentHostedOIDCConfig(),
				ExpectError: regexp.MustCompile(
					`oidc_provider: not valid on kind "hosted"`),
			},
		},
	})
}

// TestAccDataSourceDeploymentHostedAgw verifies the data source surfaces the
// derived address, which is the only way a configuration elsewhere can learn
// where a hosted gateway answers.
func TestAccDataSourceDeploymentHostedAgw(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentHostedAgwConfig(agwRegion) + `
data "hush_deployment" "hosted" {
  id = hush_deployment.hosted.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hush_deployment.hosted",
						"agw.0.region", agwRegion),
					resource.TestCheckResourceAttrPair(
						"data.hush_deployment.hosted", "agw.0.gateway_url",
						"hush_deployment.hosted", "agw.0.gateway_url"),
					// The resource hides the managed issuer, because a block
					// in state that no configuration wrote would plan a
					// removal the API refuses. The data source has no plan to
					// protect and must not report less than the API returns.
					resource.TestCheckResourceAttr("hush_deployment.hosted",
						"oidc_provider.#", "0"),
					resource.TestCheckResourceAttr("data.hush_deployment.hosted",
						"oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("data.hush_deployment.hosted",
						"oidc_provider.0.audience", oidcAudience),
				),
			},
		},
	})
}

// Regions an environment actually offers: fra everywhere, iad in production.
// The mock does not check availability, but a test that named a region no
// cluster backs would teach the wrong value.
const (
	agwRegion  = "fra"
	agwRegion2 = "iad"
)

func deploymentHostedAgwConfig(region string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "hosted" {
  name = "tf-acc-agw-hosted"
  kind = %q

  agw {
    region = %q
  }
}
`, deploymentKindHosted, region)
}

func deploymentHostedOIDCConfig() string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = "tf-acc-agw-hosted-oidc"
  kind = %q

  agw {
    region = %q
  }

  oidc_provider {
    issuer   = %q
    audience = %q
  }
}
`, deploymentKindHosted, agwRegion, oidcIssuer, oidcAudience)
}

// deploymentAgwBlockConfig renders an agw block from raw HCL, so a test can
// state a combination the helpers above cannot express.
func deploymentAgwBlockConfig(name, kind, body string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = %q
  kind = %q

  agw {
    %s
  }
}
`, name, kind, body)
}

func deploymentNoAgwConfig(name, kind string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = %q
  kind = %q
}
`, name, kind)
}
