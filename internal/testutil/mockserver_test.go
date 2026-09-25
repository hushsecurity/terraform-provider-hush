package testutil

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// A route served through Handle wins over a fixture route matching the same
// request, and a request neither describes still gets the mock's 404.
func TestHandleTakesPrecedence(t *testing.T) {
	ms := NewMockServer(&Fixtures{Endpoints: map[string]map[string]any{
		"GET /v1/things/{thing_id}": {"200": map[string]any{}},
	}})
	defer ms.Close()
	ms.Handle("GET /v1/things/{thing_id}", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]any{"id": r.PathValue("thing_id")})
	})

	for path, want := range map[string]int{
		"/v1/things/abc": http.StatusOK,
		"/v1/nothing":    http.StatusNotFound,
	} {
		resp, err := http.Get(ms.URL() + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != want {
			t.Fatalf("%s: got %d (%s), want %d", path, resp.StatusCode, body, want)
		}
		// The handler sees the path values its pattern names.
		if want == http.StatusOK && !strings.Contains(string(body), `"abc"`) {
			t.Fatalf("%s: the handler did not see thing_id: %s", path, body)
		}
	}
}
