package auth0_access_privilege

import "testing"

// A field in one and not the other only fails at apply time.
func TestSchemasCoverTheSameFields(t *testing.T) {
	resource := ResourceSchema()
	data := DataSourceSchema()

	for name := range resource {
		if data[name] == nil {
			t.Errorf("%q is in the resource schema but not the data source", name)
		}
	}
	for name := range data {
		if resource[name] == nil {
			t.Errorf("%q is in the data source schema but not the resource", name)
		}
	}
}

// The one field a practitioner must supply.
func TestApplicationIDIsRequiredOnTheResource(t *testing.T) {
	s := ResourceSchema()["application_id"]
	if s == nil {
		t.Fatal("application_id missing from the resource schema")
	}
	if !s.Required {
		t.Error("application_id must be required")
	}
	if s.Computed {
		t.Error("application_id must not be computed on the resource")
	}
	if d := DataSourceSchema()["application_id"]; !d.Computed || d.Required {
		t.Error("application_id must be computed on the data source")
	}
}
