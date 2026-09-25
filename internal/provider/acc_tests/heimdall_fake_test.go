package acc_tests

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"sync"

	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// fakeHeimdall stands in for the heimdall endpoints the fixtures leave out:
// the OpenAPI schema they are generated from hides them. It keeps only what
// the provider reads back, and refuses what heimdall refuses where a test
// depends on the refusal.
type fakeHeimdall struct {
	mu             sync.Mutex
	consentMethods map[string][]string // deployment id -> method names
	catalog        map[string]map[string]any
}

var heimdall = &fakeHeimdall{
	consentMethods: map[string][]string{},
	catalog:        fakeCatalog(),
}

func label(s string) *string { return &s }

// fakeCatalog holds one entry of each shape an application treats
// differently: several labelled addresses, a single unlabelled one that needs
// manually registered OAuth credentials, and a server Hush hosts.
func fakeCatalog() map[string]map[string]any {
	tools := []map[string]any{
		{"name": "search", "type": "read", "description": "Search", "operation": nil},
		{"name": "create", "type": "write", "description": "Create", "operation": nil},
		{"name": "drop", "type": "destructive", "description": "Drop", "operation": nil},
	}
	return map[string]map[string]any{
		"datadog": {
			"display_name": "Datadog", "category": "Observability",
			"urls": []map[string]any{
				{"label": label("US1"), "url": "https://mcp.datadoghq.com/v1/mcp"},
				{"label": label("EU"), "url": "https://mcp.datadoghq.eu/v1/mcp"},
			},
			"hosted": false, "scopes": []string{}, "tools": tools,
			"manual_registration": false, "oauth_relay": false,
		},
		"slack": {
			"display_name": "Slack", "category": "Communication & Support",
			"urls":   []map[string]any{{"label": nil, "url": "https://mcp.slack.com/mcp"}},
			"hosted": false, "scopes": []string{"chat:write", "channels:read"}, "tools": tools,
			"manual_registration": true, "oauth_relay": false,
		},
		"quickbooks": {
			"display_name": "QuickBooks", "category": "Finance",
			"urls":   []map[string]any{{"label": nil, "url": "https://quickbooks.hosted.internal/mcp"}},
			"hosted": true, "scopes": []string{"com.intuit.quickbooks.accounting"}, "tools": tools,
			"manual_registration": true, "oauth_relay": true,
		},
	}
}

func init() {
	registerMockSetup(heimdall.register)
}

func (h *fakeHeimdall) register(ms *testutil.MockServer) {
	ms.Handle("GET /v1/deployments/{deployment_id}/agw/consent_methods", h.getConsentMethods)
	ms.Handle("PUT /v1/deployments/{deployment_id}/agw/consent_methods", h.putConsentMethods)
	ms.Handle("GET /v1/applications/catalog/mcp/{app_catalog_id}", h.getCatalogEntry)
}

// The API blanks a hosted entry's addresses: they are Hush's to choose.
func (h *fakeHeimdall) getCatalogEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("app_catalog_id")
	entry, ok := h.catalog[id]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, "mcp/"+id+" not found")
		return
	}
	out := maps.Clone(entry)
	out["app_catalog_id"] = id
	out["type"] = "mcp"
	if out["hosted"] == true {
		out["urls"] = []any{}
	}
	testutil.WriteJSON(w, http.StatusOK, out)
}

func decodeBody(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

func describeConsentMethods(names []string) []map[string]string {
	out := make([]map[string]string, 0, len(names))
	for _, name := range names {
		out = append(out, map[string]string{"name": name, "description": "prompts through " + name})
	}
	return out
}

func (h *fakeHeimdall) getConsentMethods(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	names, ok := h.consentMethods[r.PathValue("deployment_id")]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, "no consent methods configured")
		return
	}
	testutil.WriteJSON(w, http.StatusOK, describeConsentMethods(names))
}

// The body is a bare, non-empty list of distinct names.
func (h *fakeHeimdall) putConsentMethods(w http.ResponseWriter, r *http.Request) {
	var names []string
	if !decodeBody(w, r, &names) {
		return
	}
	sorted := slices.Clone(names)
	slices.Sort(sorted)
	if len(names) == 0 || len(slices.Compact(sorted)) != len(names) {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "methods must be a non-empty list of distinct names")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.consentMethods[r.PathValue("deployment_id")] = names
	testutil.WriteJSON(w, http.StatusOK, describeConsentMethods(names))
}
