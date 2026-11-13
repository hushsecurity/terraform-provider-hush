package acc_tests

import (
	"maps"
	"net/http"
	"slices"
	"strconv"

	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// The resource server and resource application routes of the fake.

var (
	serverKeys = map[string]bool{
		"name": true, "display_name": true, "description": true, "audience": true,
		"scopes_supported": true, "token_lifetime": true, "token_claims": true,
	}
	resourceAppKeys = map[string]bool{
		"display_name": true, "description": true, "deployment_ids": true, "enabled": true,
		"resource": true, "scopes": true,
	}
	resourceAppChangeKeys = map[string]bool{
		"display_name": true, "description": true, "deployment_ids": true, "enabled": true,
		"scopes": true, "assignments": true, "assign_all": true,
	}
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.Handle("POST /v1/applications/resource_servers", heimdall.createServer)
		ms.Handle("POST /v1/applications/resource", heimdall.createResourceApp)
		// {a}/{b} would overlap {id}/mcp if registered as resource_servers/{name}.
		for _, method := range []string{"GET", "PATCH", "DELETE"} {
			ms.Handle(method+" /v1/applications/{a}/{b}", heimdall.resourceRoutes)
		}
	})
}

func (h *fakeHeimdall) resourceRoutes(w http.ResponseWriter, r *http.Request) {
	a, b := r.PathValue("a"), r.PathValue("b")
	h.mu.Lock()
	defer h.mu.Unlock()
	switch {
	case a == "resource_servers":
		h.serverRoute(w, r, b)
	case b == "resource" && r.Method != http.MethodDelete:
		h.resourceAppRoute(w, r, a)
	default:
		testutil.WriteError(w, http.StatusNotFound, "endpoint not found: "+r.Method+" "+r.URL.Path)
	}
}

func (h *fakeHeimdall) audienceTaken(audience any, except string) bool {
	for name, s := range h.servers {
		if name != except && s["audience"] == audience {
			return true
		}
	}
	return false
}

func (h *fakeHeimdall) createServer(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	if key := unknownKey(body, serverKeys); key != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	name, _ := body["name"].(string)
	if _, ok := h.servers[name]; ok || h.audienceTaken(body["audience"], "") {
		testutil.WriteError(w, http.StatusConflict, "the name or audience is taken")
		return
	}
	server := map[string]any{
		"description": nil, "scopes_supported": []any{}, "token_lifetime": nil, "token_claims": []any{},
	}
	maps.Copy(server, body)
	h.servers[name] = server
	testutil.WriteJSON(w, http.StatusCreated, server)
}

func (h *fakeHeimdall) serverRoute(w http.ResponseWriter, r *http.Request, name string) {
	server, ok := h.servers[name]
	if !ok {
		testutil.WriteError(w, http.StatusNotFound, "resource server "+name)
		return
	}
	switch r.Method {
	case http.MethodGet:
		testutil.WriteJSON(w, http.StatusOK, server)
	case http.MethodDelete:
		for _, app := range h.apps {
			if app["resource"] == name {
				testutil.WriteError(w, http.StatusConflict, "resource "+name+" is used by an application")
				return
			}
		}
		delete(h.servers, name)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPatch:
		var body map[string]any
		if !decodeBody(w, r, &body) {
			return
		}
		if key := unknownKey(body, serverKeys); key != "" || body["name"] != nil {
			testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
			return
		}
		if h.audienceTaken(body["audience"], name) {
			testutil.WriteError(w, http.StatusConflict, "the audience is taken")
			return
		}
		maps.Copy(server, body)
		testutil.WriteJSON(w, http.StatusOK, server)
	}
}

// scopesUnknown names a scope the resource server doesn't support.
func (h *fakeHeimdall) scopesUnknown(resource any, scopes any) string {
	var supported []string
	for _, s := range h.servers[resource.(string)]["scopes_supported"].([]any) {
		supported = append(supported, s.(map[string]any)["name"].(string))
	}
	list, _ := scopes.([]any)
	for _, scope := range list {
		if !slices.Contains(supported, scope.(string)) {
			return scope.(string)
		}
	}
	return ""
}

func (h *fakeHeimdall) createResourceApp(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	if key := unknownKey(body, resourceAppKeys); key != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.servers[body["resource"].(string)]; !ok {
		testutil.WriteError(w, http.StatusNotFound, "resource server "+body["resource"].(string))
		return
	}
	if scope := h.scopesUnknown(body["resource"], body["scopes"]); scope != "" {
		testutil.WriteError(w, http.StatusBadRequest, "no scope "+scope)
		return
	}
	h.nextAppID++
	id := "app-acc" + strconv.Itoa(h.nextAppID)
	app := map[string]any{
		"id": id, "name": body["resource"], "type": "resource", "description": nil,
		"enabled": true, "assignments": []any{}, "assign_all": false, "scopes": []any{},
	}
	maps.Copy(app, body)
	h.apps[id] = app
	testutil.WriteJSON(w, http.StatusCreated, app)
}

func (h *fakeHeimdall) resourceAppRoute(w http.ResponseWriter, r *http.Request, id string) {
	app, ok := h.apps[id]
	if !ok || app["type"] != "resource" {
		testutil.WriteError(w, http.StatusNotFound, "resource application "+id)
		return
	}
	if r.Method == http.MethodGet {
		testutil.WriteJSON(w, http.StatusOK, app)
		return
	}
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	if key := unknownKey(body, resourceAppChangeKeys); key != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
		return
	}
	if scope := h.scopesUnknown(app["resource"], body["scopes"]); scope != "" {
		testutil.WriteError(w, http.StatusBadRequest, "no scope "+scope)
		return
	}
	maps.Copy(app, body)
	testutil.WriteJSON(w, http.StatusOK, app)
}
