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

// A literal path nested under a collection is served from its fixture response
// even though the collection's {id} route also matches it, and whichever of the
// two parseRoutes happened to order first. A list endpoint keeps its envelope,
// and a fixture whose 200 is not an object is left to the store.
func TestSingletonRoutes(t *testing.T) {
	ms := NewMockServer(&Fixtures{Endpoints: map[string]map[string]any{
		"GET /v1/things/{thing_id}": {"200": map[string]any{"id": "example"}},
		"GET /v1/things/settings":   {"200": map[string]any{"token": "s3cret"}},
		"GET /v1/things":            {"200": map[string]any{"items": []any{}}},
		"GET /v1/things/logs":       {"200": []any{}},
	}})
	defer ms.Close()

	for path, want := range map[string]string{
		"/v1/things/settings": `"token":"s3cret"`,
		"/v1/things":          `"items":`,
		"/v1/things/logs":     `"items":`,
	} {
		resp, err := http.Get(ms.URL() + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: got %d (%s)", path, resp.StatusCode, body)
		}
		if !strings.Contains(string(body), want) {
			t.Fatalf("%s: got %s, want it to contain %s", path, body, want)
		}
	}
}
