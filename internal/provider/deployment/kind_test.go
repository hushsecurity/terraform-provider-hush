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
