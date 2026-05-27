package credutil

import "testing"

// fakeDiff models only what forbidDeploymentIDsChange reads from a
// *schema.ResourceDiff: the resource Id and whether deployment_ids changed.
// It does not model the attribute's value — `changed` maps directly to
// HasChange("deployment_ids"), not to any new deployment_ids value.
type fakeDiff struct {
	id      string
	changed bool
}

func (f fakeDiff) Id() string            { return f.id }
func (f fakeDiff) HasChange(string) bool { return f.changed }

func TestForbidDeploymentIDsChange(t *testing.T) {
	cases := []struct {
		name    string
		diff    fakeDiff
		wantErr bool
	}{
		// On create the resource has no Id yet, and deployment_ids going from
		// null to its configured value reports HasChange == true. The guard
		// must not fire here, otherwise every create would error.
		{"create, deployment_ids not yet set", fakeDiff{id: "", changed: false}, false},
		{"create, deployment_ids being set", fakeDiff{id: "", changed: true}, false},
		// On an existing resource: only a real change to deployment_ids is rejected.
		{"existing, deployment_ids unchanged", fakeDiff{id: "dac-1", changed: false}, false},
		{"existing, deployment_ids changed", fakeDiff{id: "dac-1", changed: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := forbidDeploymentIDsChange(tc.diff)
			if (err != nil) != tc.wantErr {
				t.Fatalf("got err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
