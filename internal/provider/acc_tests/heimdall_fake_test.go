package acc_tests

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"strconv"
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
	apps          map[string]map[string]any // application id -> application, as returned
	appSecrets    map[string]any            // application id -> client_secret
	nextAppID     int
	// Deployments whose gateway has not published its key: an enabled
	// application with a client secret cannot be placed on them.
	keyless map[string]bool
}

var heimdall = &fakeHeimdall{
	consentMethods: map[string][]string{},
	catalog:        fakeCatalog(),
	customApps:     map[string]map[string]any{},
	customSecrets:  map[string]map[string]any{},
	apps:           map[string]map[string]any{},
	appSecrets:     map[string]any{},
	keyless:        map[string]bool{},
}

func label(s string) *string { return &s }

// registerFakeCatalogEntry adds a single-address entry to the fake catalog.
func registerFakeCatalogEntry(id string, manualRegistration bool) {
	heimdall.mu.Lock()
	defer heimdall.mu.Unlock()
	if _, ok := heimdall.catalog[id]; ok {
		return
	}
	heimdall.catalog[id] = map[string]any{
		"display_name": id, "category": "Other",
		"urls":   []map[string]any{{"label": nil, "url": "https://" + id + ".example.com/mcp"}},
		"hosted": false, "scopes": []string{}, "tools": []map[string]any{},
		"manual_registration": manualRegistration, "oauth_relay": false,
	}
}

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
	ms.Handle("POST /v1/applications/custom/mcp", h.createCustomApp)
	ms.Handle("DELETE /v1/applications/custom/{name}", h.deleteCustomApp)
	ms.Handle("POST /v1/applications/mcp/{app}", h.createApp)
	ms.Handle("GET /v1/applications", h.listApps)
	ms.Handle("GET /v1/applications/{id}", h.getApp)
	ms.Handle("DELETE /v1/applications/{id}", h.deleteApp)
	ms.Handle("GET /v1/applications/{id}/mcp", h.getApp)
	ms.Handle("PATCH /v1/applications/{id}/mcp", h.patchApp)
	ms.Handle("POST /v1/applications/{id}/mcp/tool_operations", h.toolOperations)
	ms.Handle("POST /v1/applications/{id}/mcp/tool_group_operations", h.toolGroupOperations)
	// The catalog, custom app and per-entry routes all have four segments
	// under /v1/applications, and ServeMux refuses patterns that overlap
	// without one being more specific, so they share one dispatcher.
	ms.Handle("GET /v1/applications/{a}/{b}/{c}", h.dispatch(h.getCatalogEntry, h.getCustomApp, h.getApp))
	ms.Handle("PATCH /v1/applications/{a}/{b}/{c}", h.dispatch(nil, h.patchCustomApp, h.patchApp))
}

// dispatch routes /v1/applications/{a}/{b}/{c}: catalog/mcp/{id},
// custom/{name}/mcp, or {id}/mcp/{kind}. It sets the path values each handler
// reads by the name the handler expects.
func (h *fakeHeimdall) dispatch(catalog, custom, app http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, b, c := r.PathValue("a"), r.PathValue("b"), r.PathValue("c")
		switch {
		case a == "catalog" && b == "mcp" && catalog != nil:
			r.SetPathValue("app_catalog_id", c)
			catalog(w, r)
		case a == "custom" && c == "mcp":
			r.SetPathValue("name", b)
			custom(w, r)
		case b == "mcp" && appKind(c) != "":
			r.SetPathValue("id", a)
			app(w, r)
		default:
			testutil.WriteError(w, http.StatusNotFound, "endpoint not found: "+r.Method+" "+r.URL.Path)
		}
	}
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
	// Credential changes reach every application created from the custom app
	// (CustomMCPApp._cascade_changes).
	_, idChanged := body["client_id"]
	_, secretChanged := body["client_secret"]
	if idChanged || secretChanged {
		for _, id := range h.applicationsFrom("custom-" + name) {
			h.apps[id]["client_id"] = updated["client_id"]
			if secret := h.customSecrets[name]["client_secret"]; secret != nil && updated["client_id"] != nil {
				h.appSecrets[id] = secret
			} else {
				delete(h.appSecrets, id)
			}
		}
	}
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

// applicationsFrom lists the applications created from a catalog id.
func (h *fakeHeimdall) applicationsFrom(appCatalogID string) []string {
	var ids []string
	for id, app := range h.apps {
		if app["app_catalog_id"] == appCatalogID {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	return ids
}

var googleKinds = []string{"gcalendar", "gdocs", "gdrive", "gmail", "gpeople", "gsheets", "gslides"}

// appKind is the route an application's own settings travel through: its
// catalog id when it has routes of its own, "" for the generic ones.
func appKind(appCatalogID string) string {
	if slices.Contains(googleKinds, appCatalogID) || appCatalogID == "quickbooks" {
		return appCatalogID
	}
	return ""
}

// The keys each route's strict model accepts.
func appKeys(kind string, creating bool) map[string]bool {
	keys := map[string]bool{
		"display_name": true, "description": true, "deployment_ids": true, "allowed_agents": true,
		"enabled": true, "client_id": true, "client_secret": true,
	}
	if creating {
		keys["url_label"] = true
	} else {
		keys["scopes"] = true
		keys["assignments"] = true
		keys["assign_all"] = true
	}
	switch {
	case slices.Contains(googleKinds, kind):
		keys["google_project_id"] = true
	case kind == "quickbooks":
		keys["quickbooks_company_id"] = true
		keys["quickbooks_sandbox"] = true
	}
	return keys
}

func unknownKey(body map[string]any, allowed map[string]bool) string {
	for key := range body {
		if !allowed[key] {
			return key
		}
	}
	return ""
}

// entryFor resolves a catalog id to what an application copies from it.
func (h *fakeHeimdall) entryFor(appCatalogID string) (map[string]any, bool) {
	if name, ok := strings.CutPrefix(appCatalogID, "custom-"); ok {
		custom, ok := h.customApps[name]
		if !ok {
			return nil, false
		}
		return map[string]any{
			"display_name": custom["display_name"], "urls": custom["urls"], "scopes": custom["scopes"],
			"tools": custom["tools"], "hosted": false, "oauth_relay": custom["oauth_relay"],
			"manual_registration": false,
		}, true
	}
	entry, ok := h.catalog[appCatalogID]
	return entry, ok
}

func urlOptions(v any) []map[string]any {
	switch options := v.(type) {
	case []map[string]any:
		return options
	case []any:
		out := make([]map[string]any, 0, len(options))
		for _, o := range options {
			out = append(out, o.(map[string]any))
		}
		return out
	}
	return nil
}

func labelString(v any) string {
	switch label := v.(type) {
	case *string:
		return *label
	case string:
		return label
	}
	return ""
}

func (h *fakeHeimdall) createApp(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	appCatalogID := r.PathValue("app")
	kind := appKind(appCatalogID)
	if key := unknownKey(body, appKeys(kind, true)); key != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
		return
	}
	switch {
	case slices.Contains(googleKinds, kind) && body["google_project_id"] == nil:
		testutil.WriteError(w, http.StatusUnprocessableEntity, "google_project_id is required")
		return
	case kind == "quickbooks" && body["quickbooks_company_id"] == nil:
		testutil.WriteError(w, http.StatusUnprocessableEntity, "quickbooks_company_id is required")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	entry, ok := h.entryFor(appCatalogID)
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, "mcp/"+appCatalogID+" not found")
		return
	}
	for _, app := range h.apps {
		if app["display_name"] == body["display_name"] {
			testutil.WriteError(w, http.StatusConflict, "Application already exists")
			return
		}
	}
	if entry["manual_registration"] == true && (body["client_id"] == nil || body["client_secret"] == nil) {
		testutil.WriteError(w, http.StatusBadRequest, appCatalogID+" requires a manually-registered client_id and client_secret")
		return
	}

	options := urlOptions(entry["urls"])
	label, _ := body["url_label"].(string)
	name := appCatalogID
	var url any
	switch {
	case len(options) == 1 && label != "":
		testutil.WriteError(w, http.StatusUnprocessableEntity, appCatalogID+" has a single url and does not accept a label")
		return
	case len(options) == 1:
		url = options[0]["url"]
	default:
		for _, option := range options {
			if labelString(option["label"]) == label && label != "" {
				url = option["url"]
				name = appCatalogID + "-" + strings.ToLower(label)
			}
		}
		if url == nil {
			testutil.WriteError(w, http.StatusUnprocessableEntity, "url_label must be one of the entry's labels")
			return
		}
	}
	if entry["hosted"] == true {
		url = nil
	}

	h.nextAppID++
	id := "app-acc" + strconv.Itoa(h.nextAppID)
	app := map[string]any{
		"id": id, "name": name, "type": "mcp", "app_catalog_id": appCatalogID,
		"catalog_display_name": entry["display_name"],
		"description":          nil, "allowed_agents": nil, "enabled": true,
		"assignments": []any{}, "assign_all": false,
		"url": url, "hosted": entry["hosted"], "scopes": entry["scopes"],
		"oauth_relay": entry["oauth_relay"], "client_id": nil,
		"tool_groups": []any{
			map[string]any{"type": "read", "operation": "allow"},
			map[string]any{"type": "write", "operation": "user_consent"},
			map[string]any{"type": "destructive", "operation": "block"},
		},
		"tools": cloneTools(entry["tools"]),
	}
	delete(body, "url_label")
	// An application of a custom app that names no OAuth client takes the
	// custom app's (catalog_resolver.get_manual_credentials).
	if name, custom := strings.CutPrefix(appCatalogID, "custom-"); custom && body["client_id"] == nil {
		if clientID := h.customApps[name]["client_id"]; clientID != nil {
			app["client_id"] = clientID
			if secret := h.customSecrets[name]["client_secret"]; secret != nil {
				h.appSecrets[id] = secret
			}
		}
	}
	if problem := h.setAppFields(id, app, body); problem != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	// Saved before the key is looked for, as heimdall does.
	h.apps[id] = app
	if h.needsMissingKey(id, app) {
		testutil.WriteError(w, http.StatusServiceUnavailable, noPublicKey)
		return
	}
	testutil.WriteJSON(w, http.StatusCreated, h.appOut(app, kind))
}

// cloneTools copies an entry's tools, so an application's operations stay
// its own.
func cloneTools(v any) []any {
	var out []any
	raw, _ := json.Marshal(v)
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		out = []any{}
	}
	return out
}

// needsMissingKey reports whether an application would have its secret
// encrypted to a gateway that has no key to encrypt it to.
func (h *fakeHeimdall) needsMissingKey(id string, app map[string]any) bool {
	if app["enabled"] != true || h.appSecrets[id] == nil {
		return false
	}
	deployments, _ := app["deployment_ids"].([]any)
	for _, dep := range deployments {
		if h.keyless[dep.(string)] {
			return true
		}
	}
	return false
}

const noPublicKey = "Service unavailable: deployment has no public key; not ready to accept encrypted parameters"

func (h *fakeHeimdall) setAppFields(id string, app, body map[string]any) string {
	for key, value := range body {
		switch key {
		case "client_secret":
			if value == nil {
				delete(h.appSecrets, id)
			} else {
				h.appSecrets[id] = value
			}
		case "allowed_agents":
			if list, ok := value.([]any); ok && len(list) == 0 {
				return "allowed_agents must not be empty"
			}
			app[key] = value
		case "quickbooks_company_id", "quickbooks_sandbox", "google_project_id":
			if value != nil {
				app[key] = value
			}
		default:
			app[key] = value
		}
	}
	return ""
}

// appOut is an application as a route returns it: the generic routes know
// nothing of the Google and QuickBooks settings.
func (h *fakeHeimdall) appOut(app map[string]any, kind string) map[string]any {
	out := maps.Clone(app)
	switch kind {
	case "":
		delete(out, "google_project_id")
		delete(out, "quickbooks_company_id")
		delete(out, "quickbooks_sandbox")
	case "quickbooks":
		if _, ok := out["quickbooks_sandbox"]; !ok {
			out["quickbooks_sandbox"] = false
		}
	}
	return out
}

// routeKind is the kind a request's path names, which must be the
// application's own.
func routeKind(r *http.Request) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 5 {
		return parts[4]
	}
	return ""
}

func (h *fakeHeimdall) fetchApp(w http.ResponseWriter, r *http.Request) (map[string]any, string, bool) {
	app, ok := h.apps[r.PathValue("id")]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, r.PathValue("id")+" not found")
		return nil, "", false
	}
	kind := routeKind(r)
	if kind != "" && kind != app["app_catalog_id"] {
		testutil.WriteError(w, http.StatusBadRequest, "application is not a "+kind+" application")
		return nil, "", false
	}
	return app, kind, true
}

func (h *fakeHeimdall) getApp(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if app, kind, ok := h.fetchApp(w, r); ok {
		testutil.WriteJSON(w, http.StatusOK, h.appOut(app, kind))
	}
}

func (h *fakeHeimdall) patchApp(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	app, kind, ok := h.fetchApp(w, r)
	if !ok {
		return
	}
	// The settings an entry's own route carries are refused anywhere else,
	// so a generic patch of such an application would silently lose them.
	if key := unknownKey(body, appKeys(kind, false)); key != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
		return
	}
	if kind == "" && appKind(app["app_catalog_id"].(string)) != "" {
		testutil.WriteError(w, http.StatusBadRequest, "use the application's own route")
		return
	}
	updated := maps.Clone(app)
	if problem := h.setAppFields(r.PathValue("id"), updated, body); problem != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	// Stored before the key is looked for, as heimdall does.
	h.apps[r.PathValue("id")] = updated
	if h.needsMissingKey(r.PathValue("id"), updated) {
		testutil.WriteError(w, http.StatusServiceUnavailable, noPublicKey)
		return
	}
	testutil.WriteJSON(w, http.StatusOK, h.appOut(updated, kind))
}

func (h *fakeHeimdall) deleteApp(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	app, ok := h.apps[r.PathValue("id")]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, r.PathValue("id")+" not found")
		return
	}
	delete(h.apps, r.PathValue("id"))
	delete(h.appSecrets, r.PathValue("id"))
	testutil.WriteJSON(w, http.StatusOK, app)
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

// toolOperations sets tools' own operations by name; an unknown name is a
// 404, and nothing changes.
func (h *fakeHeimdall) toolOperations(w http.ResponseWriter, r *http.Request) {
	var changes []map[string]any
	if !decodeBody(w, r, &changes) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	app, ok := h.apps[r.PathValue("id")]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, r.PathValue("id")+" not found")
		return
	}
	tools := cloneTools(app["tools"])
	byName := map[string]map[string]any{}
	for _, tool := range tools {
		byName[tool.(map[string]any)["name"].(string)] = tool.(map[string]any)
	}
	for _, change := range changes {
		tool, ok := byName[change["name"].(string)]
		if !ok {
			testutil.WriteError(w, http.StatusNotFound, change["name"].(string)+" not found")
			return
		}
		tool["operation"] = change["operation"]
	}
	app["tools"] = tools
	testutil.WriteJSON(w, http.StatusOK, h.appOut(app, ""))
}

func (h *fakeHeimdall) toolGroupOperations(w http.ResponseWriter, r *http.Request) {
	var changes []map[string]any
	if !decodeBody(w, r, &changes) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	app, ok := h.apps[r.PathValue("id")]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, r.PathValue("id")+" not found")
		return
	}
	groups := cloneTools(app["tool_groups"])
	for _, change := range changes {
		if change["operation"] == nil {
			testutil.WriteError(w, http.StatusUnprocessableEntity, "a group needs an operation")
			return
		}
		for _, group := range groups {
			if group.(map[string]any)["type"] == change["type"] {
				group.(map[string]any)["operation"] = change["operation"]
			}
		}
	}
	app["tool_groups"] = groups
	testutil.WriteJSON(w, http.StatusOK, h.appOut(app, ""))
}

// listApps serves the summary list, filtered by type, one application per
// page: the provider has to follow the cursor to find anything past the first.
func (h *fakeHeimdall) listApps(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ids := make([]string, 0, len(h.apps))
	for id, app := range h.apps {
		if typ := r.URL.Query().Get("type"); typ == "" || app["type"] == typ {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	start, _ := strconv.Atoi(r.URL.Query().Get("cursor"))
	items := []any{}
	var next any
	if start < len(ids) {
		app := h.apps[ids[start]]
		items = append(items, map[string]any{
			"id": app["id"], "name": app["name"], "type": app["type"],
			"app_catalog_id": app["app_catalog_id"], "display_name": app["display_name"],
		})
		if start+1 < len(ids) {
			next = strconv.Itoa(start + 1)
		}
	}
	testutil.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "next_page": next})
}
