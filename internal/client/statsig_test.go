package client

import (
	"encoding/json"
	"testing"
)

func marshalJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestStatsigPrivilegeCreateSendsUnsetScopingAsNull(t *testing.T) {
	got := marshalJSON(t, &CreateStatsigAccessPrivilegeInput{
		Name: "n",
		Key:  StatsigKey{KeyType: "server"},
	})
	want := `{"environments":null,"key_type":"server","name":"n",` +
		`"secondary_target_app_ids":null,"target_app_id":null}`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestStatsigPrivilegeCreateSendsTheScoping(t *testing.T) {
	got := marshalJSON(t, &CreateStatsigAccessPrivilegeInput{
		Name:        "n",
		Description: "d",
		Key: StatsigKey{
			KeyType:               "client",
			Environments:          []string{"production"},
			TargetAppID:           "app-1",
			SecondaryTargetAppIDs: []string{"app-2"},
		},
	})
	want := `{"description":"d","environments":["production"],"key_type":"client",` +
		`"name":"n","secondary_target_app_ids":["app-2"],"target_app_id":"app-1"}`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

// A rename leaves the key alone; a key change sends every field, so one
// removed from the configuration is cleared.
func TestStatsigPrivilegeUpdate(t *testing.T) {
	name := "renamed"
	if got := marshalJSON(t, &UpdateStatsigAccessPrivilegeInput{Name: &name}); got != `{"name":"renamed"}` {
		t.Fatalf("rename sent %s", got)
	}

	got := marshalJSON(t, &UpdateStatsigAccessPrivilegeInput{
		Key: &StatsigKey{KeyType: "server", TargetAppID: "app-1"},
	})
	want := `{"environments":null,"key_type":"server",` +
		`"secondary_target_app_ids":null,"target_app_id":"app-1"}`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
