package client

import "testing"

func TestCloudinaryPrivilegeCreateSendsTheOtherGrantAsNull(t *testing.T) {
	got := marshalJSON(t, &CreateCloudinaryAccessPrivilegeInput{
		Name:  "n",
		Grant: CloudinaryGrant{CloudName: "product1", RoleID: "cld::role::prodenv::editor"},
	})
	want := `{"cloud_name":"product1","name":"n","role_id":"cld::role::prodenv::editor",` +
		`"system_policy_ids":null}`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

// A rename leaves the grant alone; a grant change sends every field, so the
// one switched away from is cleared in the same request.
func TestCloudinaryPrivilegeUpdate(t *testing.T) {
	name := "renamed"
	if got := marshalJSON(t, &UpdateCloudinaryAccessPrivilegeInput{Name: &name}); got != `{"name":"renamed"}` {
		t.Fatalf("rename sent %s", got)
	}

	got := marshalJSON(t, &UpdateCloudinaryAccessPrivilegeInput{
		Description: new(string),
		Grant: &CloudinaryGrant{
			CloudName:       "product1",
			SystemPolicyIDs: []string{"cld::policy::global::a", "cld::policy::global::b"},
		},
	})
	want := `{"cloud_name":"product1","description":"","role_id":null,` +
		`"system_policy_ids":["cld::policy::global::a","cld::policy::global::b"]}`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
