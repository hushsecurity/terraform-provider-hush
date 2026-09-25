package acc_tests

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"strings"
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
	customApps     map[string]map[string]any // name -> custom app, as returned
	// Secrets the API never echoes, kept so a test can see what was sent.
	customSecrets map[string]map[string]any // name -> client_secret, auth secret
}

var heimdall = &fakeHeimdall{
	consentMethods: map[string][]string{},
	catalog:        fakeCatalog(),
	customApps:     map[string]map[string]any{},
	customSecrets:  map[string]map[string]any{},
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
	ms.Handle("POST /v1/applications/custom/mcp", h.createCustomApp)
	ms.Handle("GET /v1/applications/custom/{name}/mcp", h.getCustomApp)
	ms.Handle("PATCH /v1/applications/custom/{name}/mcp", h.patchCustomApp)
	ms.Handle("DELETE /v1/applications/custom/{name}", h.deleteCustomApp)
}

// Every key a custom app's create or patch may carry: anything else is
// refused, as heimdall's strict models refuse it.
var customAppKeys = map[string]bool{
	"name": true, "display_name": true, "description": true, "urls": true, "scopes": true,
	"tools": true, "oauth_relay": true, "headers": true, "client_id": true,
	"client_secret": true, "auth": true,
}

var authSecretKey = map[string]string{"bearer": "token", "basic": "password", "header": "value"}

// setCustomAppFields applies a create or patch body. The secrets are split
// off, as heimdall keeps them apart, and auth is masked the way it is shown.
func (h *fakeHeimdall) setCustomAppFields(name string, app, body map[string]any) string {
	for key := range body {
		if !customAppKeys[key] {
			return "unknown field " + key
		}
	}
	secrets := h.customSecrets[name]
	for key, value := range body {
		switch key {
		case "name":
		case "client_secret":
			secrets["client_secret"] = value
		case "client_id":
			app["client_id"] = value
			if value == nil {
				delete(secrets, "client_secret")
			}
		case "auth":
			if value == nil {
				app["auth"] = nil
				delete(secrets, "auth")
				continue
			}
			auth := maps.Clone(value.(map[string]any))
			field := authSecretKey[auth["type"].(string)]
			if auth[field] == nil || auth[field] == "" {
				return "auth requires its secret"
			}
			secrets["auth"] = auth[field]
			auth[field] = "****"
			app["auth"] = auth
		case "headers":
			if value == nil {
				return "headers cannot be null"
			}
			app["headers"] = value
		default:
			app[key] = value
		}
	}
	return ""
}

func (h *fakeHeimdall) createCustomApp(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	name, _ := body["name"].(string)
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.customApps[name]; exists {
		testutil.WriteError(w, http.StatusConflict, name+" already exists")
		return
	}
	app := map[string]any{
		"name": name, "app_catalog_id": "custom-" + name, "type": "mcp", "description": nil,
		"scopes": []any{}, "tools": []any{}, "oauth_relay": false, "headers": map[string]any{},
		"client_id": nil, "auth": nil,
	}
	h.customSecrets[name] = map[string]any{}
	if problem := h.setCustomAppFields(name, app, body); problem != "" {
		delete(h.customSecrets, name)
		testutil.WriteError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	h.customApps[name] = app
	testutil.WriteJSON(w, http.StatusCreated, app)
}

func (h *fakeHeimdall) getCustomApp(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	app, ok := h.customApps[r.PathValue("name")]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, r.PathValue("name")+" not found")
		return
	}
	testutil.WriteJSON(w, http.StatusOK, app)
}

func (h *fakeHeimdall) patchCustomApp(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	name := r.PathValue("name")
	h.mu.Lock()
	defer h.mu.Unlock()
	app, ok := h.customApps[name]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, name+" not found")
		return
	}
	if _, ok := body["name"]; ok {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field name")
		return
	}
	updated := maps.Clone(app)
	if problem := h.setCustomAppFields(name, updated, body); problem != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	h.customApps[name] = updated
	testutil.WriteJSON(w, http.StatusOK, updated)
}

func (h *fakeHeimdall) deleteCustomApp(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	h.mu.Lock()
	defer h.mu.Unlock()
	app, ok := h.customApps[name]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, name+" not found")
		return
	}
	if users := h.applicationsFrom("custom-" + name); len(users) > 0 {
		testutil.WriteError(w, http.StatusConflict, name+" is used by "+strings.Join(users, ", "))
		return
	}
	delete(h.customApps, name)
	delete(h.customSecrets, name)
	testutil.WriteJSON(w, http.StatusOK, app)
}

// applicationsFrom lists the applications created from a catalog id. None
// exist yet; the MCP applications fake fills this in.
func (h *fakeHeimdall) applicationsFrom(string) []string { return nil }

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
