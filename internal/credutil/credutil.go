// Package credutil holds small helpers shared across access credential resources.
package credutil

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// diffChecker is the subset of *schema.ResourceDiff that
// ForbidDeploymentIDsChange needs; it keeps the logic unit-testable.
type diffChecker interface {
	Id() string
	HasChange(string) bool
}

// ForbidDeploymentIDsChange is a CustomizeDiff that rejects changes to
// deployment_ids on an already-created credential.
//
// The Hush API does not accept deployment_ids on update for these credential
// types (only the workload-identity-federation types do); its update model is
// strict and rejects the field. Allowing the change in Terraform would either
// silently drift (the field is dropped) or fail mid-apply with a 422. Rejecting
// it at plan time gives a clear, early error: the credential must be recreated
// to change its deployment set.
func ForbidDeploymentIDsChange(_ context.Context, d *schema.ResourceDiff, _ any) error {
	return forbidDeploymentIDsChange(d)
}

func forbidDeploymentIDsChange(d diffChecker) error {
	if d.Id() != "" && d.HasChange("deployment_ids") {
		return fmt.Errorf("deployment_ids cannot be changed after creation; " +
			"delete and recreate the resource to change its deployment set")
	}
	return nil
}
