package client

import (
	"encoding/json"
	"testing"
)

// buckets has three update states: omitted when unchanged, null for every
// bucket (which the Hush API reads apart from omitted), and the list.
func TestUpdateCloudflareR2AccessPrivilegeInput_BucketsMarshaling(t *testing.T) {
	cases := []struct {
		name  string
		input UpdateCloudflareR2AccessPrivilegeInput
		want  string
	}{
		{"unchanged", UpdateCloudflareR2AccessPrivilegeInput{}, `{}`},
		{"every bucket", UpdateCloudflareR2AccessPrivilegeInput{Buckets: NewBucketsUpdate(nil)}, `{"buckets":null}`},
		{
			"listed",
			UpdateCloudflareR2AccessPrivilegeInput{Buckets: NewBucketsUpdate([]string{"exports", "photos"})},
			`{"buckets":["exports","photos"]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

// Create always sends buckets: the Hush API refuses an object permission
// without them, and null is every bucket.
func TestCreateCloudflareR2AccessPrivilegeInput_SendsNullBuckets(t *testing.T) {
	got, err := json.Marshal(CreateCloudflareR2AccessPrivilegeInput{
		Name: "n", Permission: "object_read_only", Jurisdiction: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"name":"n","permission":"object_read_only","buckets":null,"jurisdiction":"default"}`
	if string(got) != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
