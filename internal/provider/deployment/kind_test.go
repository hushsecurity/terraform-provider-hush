package deployment

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// kindDiff runs the real resource diff, CustomizeDiff included, for a
// deployment that exists and whose configuration now names another kind.
func kindDiff(t *testing.T, stateKind, configKind string) error {
	t.Helper()
	state := &terraform.InstanceState{
		ID: "dep-1",
		Attributes: map[string]string{
			"id":   "dep-1",
			"name": "d",
			"kind": stateKind,
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name": "d",
		"kind": configKind,
	})
	_, err := Resource().Diff(context.Background(), state, config, nil)
	return err
}

// The API fixes the kind at creation, so the provider must not plan an update
// it cannot complete. Refusing costs a plan; the 422 it replaces cost an apply
// and left the configuration proposing the same failure on every run.
func TestKindCannotBeEditedOnAnExistingDeployment(t *testing.T) {
	err := kindDiff(t, "k8s", "ecs")
	if err == nil {
		t.Fatal("expected the kind change to be refused")
	}
	for _, want := range []string{"cannot be changed", `"k8s"`, `"ecs"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected the refusal to mention %s, got: %v", want, err)
		}
	}
}

// The same kind is not a change, so an unrelated edit still plans.
func TestKindUnchangedIsNotAChange(t *testing.T) {
	if err := kindDiff(t, "k8s", "k8s"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// A create has no prior kind to move away from.
func TestKindOnCreateIsNotAChange(t *testing.T) {
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name": "d", "kind": "ecs",
	})
	if _, err := Resource().Diff(
		context.Background(), nil, config, nil,
	); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// A deployment created before the kind was mandatory carries none, and the API
// refuses to set one on a deployment that exists. Skipping it would leave the
// field unwritable and unmentioned: the plan would offer the change forever and
// each apply would report success without sending a request.
func TestKindAbsentOnTheDeploymentIsRefused(t *testing.T) {
	err := kindDiff(t, "", "k8s")
	if err == nil {
		t.Fatal("expected a deployment with no kind to be refused")
	}
	for _, want := range []string{"records none", `"k8s"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected the refusal to mention %s, got: %v", want, err)
		}
	}
}

// The refusal has to name steps that work. A replacement keeps the prior state,
// so this rule fires during one too, which makes "terraform apply -replace" the
// wrong thing to send a reader towards.
func TestKindRefusalNamesAPathThatWorks(t *testing.T) {
	err := kindDiff(t, "k8s", "ecs")
	if err == nil {
		t.Fatal("expected the kind change to be refused")
	}
	if strings.Contains(err.Error(), "-replace") {
		t.Fatalf("the refusal must not recommend -replace, which it blocks: %v", err)
	}
	if !strings.Contains(err.Error(), "remove this resource from the configuration") {
		t.Fatalf("expected the refusal to name the steps that work, got: %v", err)
	}
}

// kind is deliberately not ForceNew: a replacement reissues the deployment's
// credentials and detaches the applications bound to its gateway, so a change
// is refused rather than answered with a destroy. No diff test would notice
// this being lost, because CustomizeDiff runs either way.
func TestKindIsNotForceNew(t *testing.T) {
	if DeploymentResourceSchema()["kind"].ForceNew {
		t.Fatal("kind must not be ForceNew: a change is refused, not replaced")
	}
}

// systemd is a Linux VM the sensor or the connector is installed on, and the
// API takes it; the schema must not refuse it before the request is sent, and
// must still refuse a kind the API has never heard of.
func TestKindSchemaAcceptsSystemd(t *testing.T) {
	validate := DeploymentResourceSchema()["kind"].ValidateFunc
	for _, kind := range []string{"k8s", "hosted", "ecs", "serverless", "systemd"} {
		if _, errs := validate(kind, "kind"); len(errs) > 0 {
			t.Errorf("expected kind %q to be accepted, got %v", kind, errs)
		}
	}
	if _, errs := validate("vm", "kind"); len(errs) == 0 {
		t.Error(`expected kind "vm" to be refused`)
	}
}

// A systemd deployment is created like any other kind.
func TestKindSystemdOnCreateIsNotAChange(t *testing.T) {
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name": "d", "kind": "systemd",
	})
	if _, err := Resource().Diff(
		context.Background(), nil, config, nil,
	); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// and is fixed at creation like any other.
func TestKindSystemdCannotBeEdited(t *testing.T) {
	err := kindDiff(t, "systemd", "k8s")
	if err == nil {
		t.Fatal("expected the kind change to be refused")
	}
	for _, want := range []string{"cannot be changed", `"systemd"`, `"k8s"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected the refusal to mention %s, got: %v", want, err)
		}
	}
}
