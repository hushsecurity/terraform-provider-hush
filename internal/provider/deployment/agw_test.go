package deployment

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

// A block carries the hostname through to the request unchanged: the API
// rejects mixed case rather than normalizing it, so anything the provider did
// to the value here would show up as drift.
func TestExpandAgw(t *testing.T) {
	d := resourceDataFor(t, map[string]any{
		"agw": []any{map[string]any{"hostname": "gw.example.com"}},
	})

	got := expandAgw(d)
	if got == nil {
		t.Fatal("expected an agw object, got nil")
	}
	if got.Hostname != "gw.example.com" {
		t.Fatalf("unexpected hostname %q", got.Hostname)
	}
	if got.Region != "" {
		t.Fatalf("a gateway the customer runs has no region, got %q", got.Region)
	}
}

// The member the kind does not use must stay empty, or `omitempty` sends the
// field the API refuses for that kind.
func TestExpandAgwHosted(t *testing.T) {
	d := resourceDataFor(t, map[string]any{
		"agw": []any{map[string]any{"region": "fra"}},
	})

	got := expandAgw(d)
	if got == nil {
		t.Fatal("expected an agw object, got nil")
	}
	if got.Region != "fra" {
		t.Fatalf("unexpected region %q", got.Region)
	}
	if got.Hostname != "" {
		t.Fatalf("a hosted gateway has no hostname, got %q", got.Hostname)
	}
}

// gateway_url is derived and the request model has no field for it, so a read
// that put it in state must not send it back.
func TestExpandAgwDropsTheDerivedURL(t *testing.T) {
	d := resourceDataFor(t, map[string]any{
		"agw": []any{map[string]any{
			"hostname":    "gw.example.com",
			"gateway_url": "https://gw.example.com/mcp",
		}},
	})

	if got := expandAgw(d).GatewayURL; got != "" {
		t.Fatalf("expected the derived url to be dropped, got %q", got)
	}
}

// An update may carry the hostname and nothing else: a region is fixed when the
// gateway is placed, and the update model has no field for it.
func TestExpandAgwUpdateSendsHostnameOnly(t *testing.T) {
	d := resourceDataFor(t, map[string]any{
		"agw": []any{map[string]any{"hostname": "gw.example.com", "region": "fra"}},
	})

	got := expandAgwUpdate(d)
	if got == nil {
		t.Fatal("expected an agw object, got nil")
	}
	if got.Hostname != "gw.example.com" {
		t.Fatalf("unexpected hostname %q", got.Hostname)
	}
	if got.Region != "" {
		t.Fatalf("expected no region on an update, got %q", got.Region)
	}
}

// No block must expand to nil, so an update that dropped it sends null and
// removes the whole object, and a create without one sends nothing.
func TestExpandAgwAbsent(t *testing.T) {
	d := resourceDataFor(t, map[string]any{})
	if got := expandAgw(d); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// Both kinds of gateway read back through the same block, each leaving the
// member that does not apply to it empty.
func TestFlattenAgw(t *testing.T) {
	tests := map[string]struct {
		agw                          *client.Agw
		hostname, region, gatewayURL string
	}{
		"customer-run": {
			agw: &client.Agw{
				Hostname:   "gw.example.com",
				GatewayURL: "https://gw.example.com/mcp",
			},
			hostname:   "gw.example.com",
			gatewayURL: "https://gw.example.com/mcp",
		},
		"hush-hosted": {
			agw: &client.Agw{
				Region:     "fra",
				GatewayURL: "https://dep-1.agw.fra.example.com/mcp",
			},
			region:     "fra",
			gatewayURL: "https://dep-1.agw.fra.example.com/mcp",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := flattenAgw(&client.Deployment{Agw: tc.agw})
			if len(got) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(got))
			}
			if got[0]["hostname"] != tc.hostname {
				t.Fatalf("unexpected hostname %v", got[0]["hostname"])
			}
			if got[0]["region"] != tc.region {
				t.Fatalf("unexpected region %v", got[0]["region"])
			}
			if got[0]["gateway_url"] != tc.gatewayURL {
				t.Fatalf("unexpected gateway_url %v", got[0]["gateway_url"])
			}
		})
	}
}

// A deployment with no gateway reads back as no block -- an empty list rather
// than null, or it would show a permanent diff against a configuration that
// declares none.
func TestFlattenAgwWithoutGateway(t *testing.T) {
	got := flattenAgw(&client.Deployment{})
	if got == nil {
		t.Fatal("expected an empty list, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected no entries, got %d", len(got))
	}
}

// The API fills the OIDC issuer in itself on a hosted deployment and then
// refuses to be sent either field. Reading it back would put a block in state
// that no configuration wrote, and the plan to remove it could never apply.
func TestFlattenOidcProvidersSkipsHosted(t *testing.T) {
	got := flattenOidcProviders(&client.Deployment{
		Kind: deploymentKindHosted,
		OidcProvider: &client.OidcConfig{
			Issuer:   "https://oidc.eks.eu-west-1.amazonaws.com/id/AAAA",
			Audience: "https://kubernetes.default.svc",
		},
	}, true)
	if len(got) != 0 {
		t.Fatalf("expected the managed provider to be hidden, got %v", got)
	}
}

// The plan-time rules mirror the hush-agw chart's, so a hostname the chart
// refuses is refused here too, and the reason says which rule was broken.
func TestValidateGatewayHostname(t *testing.T) {
	tests := []struct {
		host    string
		wantErr string
	}{
		{host: "gw.example.com"},
		{host: "mcp-gateway.eu.example.co.uk"},
		{host: "gw1.example.com"},
		{host: "GW.example.com", wantErr: "must be lowercase"},
		{host: "gw.Example.com", wantErr: "must be lowercase"},
		{host: "https://gw.example.com", wantErr: "bare domain name"},
		{host: "gw.example.com:8443", wantErr: "bare domain name"},
		{host: "gw.example.com/mcp", wantErr: "bare domain name"},
		{host: "localhost", wantErr: "bare domain name"},
		{host: "-gw.example.com", wantErr: "bare domain name"},
		{host: "gw..example.com", wantErr: "bare domain name"},
		{host: "", wantErr: "bare domain name"},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			_, errs := validateGatewayHostname(tc.host, "agw.0.hostname")
			if tc.wantErr == "" {
				if len(errs) != 0 {
					t.Fatalf("expected %q to be accepted, got %v", tc.host, errs)
				}
				return
			}
			if len(errs) != 1 {
				t.Fatalf("expected %q to be refused once, got %v", tc.host, errs)
			}
			if !strings.Contains(errs[0].Error(), tc.wantErr) {
				t.Fatalf("expected the reason for %q to mention %q, got %v",
					tc.host, tc.wantErr, errs[0])
			}
		})
	}
}

// The block holds one gateway, matching the single agw object the API stores.
// Neither member can be Required, because which one applies is decided by the
// kind; validateAgwForKind is what demands the right one, and the tests of the
// plan-time rules cover that.
func TestAgwBlockShape(t *testing.T) {
	block, ok := DeploymentResourceSchema()["agw"]
	if !ok {
		t.Fatal("agw missing from the resource schema")
	}
	if block.MaxItems != 1 {
		t.Fatalf("expected the block capped at 1, got %d", block.MaxItems)
	}
	if !block.Optional {
		t.Fatal("agw must be optional: most deployments run no gateway")
	}

	members := block.Elem.(*schema.Resource).Schema
	for _, name := range []string{"hostname", "region"} {
		member, ok := members[name]
		if !ok {
			t.Fatalf("%s missing from the agw block", name)
		}
		if member.Required {
			t.Fatalf("%s cannot be required: it applies to one kind only", name)
		}
	}
	if members["hostname"].ForceNew {
		t.Fatal("a hostname follows a DNS move in place, the API allows it")
	}
	// A region is set once, but ForceNew would answer that with a destroy, and
	// destroying the deployment detaches every application bound to the
	// gateway. validateRegionImmutable refuses the change instead.
	if members["region"].ForceNew {
		t.Fatal("region must not be ForceNew: a change is refused, not replaced")
	}
	if !members["gateway_url"].Computed {
		t.Fatal("gateway_url is derived by the API")
	}
}

// hostedDiff runs the real resource diff, CustomizeDiff included, for a hosted
// deployment moving from one region to another.
func hostedDiff(t *testing.T, stateRegion, configRegion string) error {
	t.Helper()
	state := &terraform.InstanceState{
		ID: "dep-1",
		Attributes: map[string]string{
			"id":             "dep-1",
			"name":           "gw",
			"kind":           deploymentKindHosted,
			"agw.#":          "1",
			"agw.0.region":   stateRegion,
			"agw.0.hostname": "",
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name": "gw",
		"kind": deploymentKindHosted,
		"agw":  []any{map[string]any{"region": configRegion}},
	})
	_, err := Resource().Diff(context.Background(), state, config, nil)
	return err
}

// Editing the region of a gateway that exists is refused outright. The API has
// no field for it, and answering with a replacement would detach every
// application from the gateway by way of the deployment id.
func TestRegionCannotBeEditedOnAnExistingDeployment(t *testing.T) {
	err := hostedDiff(t, "fra", "iad")
	if err == nil {
		t.Fatal("expected the move to be refused")
	}
	for _, want := range []string{"cannot be changed", `"fra"`, `"iad"`, "detaches"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected the refusal to mention %s, got: %v", want, err)
		}
	}
}

// The same region is not a move, so a plan that touches nothing else is clean.
func TestRegionUnchangedIsNotAMove(t *testing.T) {
	if err := hostedDiff(t, "fra", "fra"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// unknownValue is the sentinel Terraform puts in a plan for a value that is
// only settled during the apply -- a hostname taken from the DNS record that
// publishes it, say. ResourceDiff reads it back as the empty string, so the
// plan-time rules have to recognise it as a value the caller did write.
const unknownValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

// planFor runs the real resource diff for a fresh deployment, CustomizeDiff
// included, so the rules are exercised the way Terraform reaches them.
func planFor(t *testing.T, kind string, agw map[string]any, extra map[string]any) error {
	t.Helper()
	raw := map[string]any{"name": "d", "kind": kind}
	if agw != nil {
		raw["agw"] = []any{agw}
	}
	for k, v := range extra {
		raw[k] = v
	}
	_, err := Resource().Diff(
		context.Background(), nil, terraform.NewResourceConfigRaw(raw), nil)
	return err
}

func TestAgwKindRules(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		agw     map[string]any
		extra   map[string]any
		wantErr string
	}{
		{name: "k8s with a hostname", kind: "k8s",
			agw: map[string]any{"hostname": "gw.example.com"}},
		{name: "k8s with no block at all", kind: "k8s"},
		{name: "k8s without a hostname", kind: "k8s",
			agw: map[string]any{}, wantErr: `agw.hostname: required on kind "k8s"`},
		{name: "k8s with a region", kind: "k8s",
			agw:     map[string]any{"hostname": "gw.example.com", "region": "fra"},
			wantErr: `agw.region: not valid on kind "k8s"`},
		{name: "hosted with a region", kind: deploymentKindHosted,
			agw: map[string]any{"region": "fra"}},
		{name: "hosted without a region", kind: deploymentKindHosted,
			agw:     map[string]any{},
			wantErr: `agw.region: required on kind "hosted"`},
		{name: "hosted with no block at all", kind: deploymentKindHosted,
			wantErr: `agw: an agw block with a region is required on kind "hosted"`},
		{name: "hosted with a hostname", kind: deploymentKindHosted,
			agw:     map[string]any{"region": "fra", "hostname": "gw.example.com"},
			wantErr: `agw.hostname: not valid on kind "hosted"`},
		{name: "hosted with an oidc_provider", kind: deploymentKindHosted,
			agw: map[string]any{"region": "fra"},
			extra: map[string]any{"oidc_provider": []any{map[string]any{
				"issuer": "https://issuer.example.com", "audience": "hush"}}},
			wantErr: `oidc_provider: not valid on kind "hosted"`},
		{name: "a kind that cannot run a gateway", kind: "ecs",
			agw:     map[string]any{"hostname": "gw.example.com"},
			wantErr: `agw: a gateway is only valid on kind "k8s" or "hosted"`},

		// A member the caller wrote but Terraform cannot resolve yet must not
		// be mistaken for one they left out, or the ordinary shape of this
		// feature -- the gateway's DNS record and the deployment in one
		// configuration -- never reaches an apply.
		{name: "k8s with an unknown hostname", kind: "k8s",
			agw: map[string]any{"hostname": unknownValue}},
		{name: "hosted with an unknown region", kind: deploymentKindHosted,
			agw: map[string]any{"region": unknownValue}},
		// Still written, so the rule that refuses it still applies.
		{name: "hosted with an unknown hostname", kind: deploymentKindHosted,
			agw:     map[string]any{"region": "fra", "hostname": unknownValue},
			wantErr: `agw.hostname: not valid on kind "hosted"`},
		{name: "k8s with an unknown region", kind: "k8s",
			agw:     map[string]any{"hostname": "gw.example.com", "region": unknownValue},
			wantErr: `agw.region: not valid on kind "k8s"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := planFor(t, tc.kind, tc.agw, tc.extra)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected the plan to be accepted, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected %q, got no error", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected an error mentioning %q, got %v", tc.wantErr, err)
			}
		})
	}
}
