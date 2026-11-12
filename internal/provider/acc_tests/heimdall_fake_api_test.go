package acc_tests

import (
	"maps"
	"net/http"

	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// The custom api app routes of the fake.

var customAPIAppKeys = map[string]bool{
	"name": true, "display_name": true, "description": true, "url": true, "spec_url": true,
	"headers": true, "auth": true,
}

func (h *fakeHeimdall) createCustomAPIApp(w http.ResponseWriter, r *http.Request) {
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
		"name": name, "app_catalog_id": "custom-" + name, "type": "api", "description": nil,
		"spec_url": nil, "detect_required": false, "headers": map[string]any{}, "auth": nil,
	}
	h.customSecrets[name] = map[string]any{}
	if problem := h.setCustomAppFields(name, app, body, customAPIAppKeys); problem != "" {
		delete(h.customSecrets, name)
		testutil.WriteError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	h.customApps[name] = app
	testutil.WriteJSON(w, http.StatusCreated, app)
}

// patchCustomAPIApp flags a changed url or spec_url for the next detect.
func (h *fakeHeimdall) patchCustomAPIApp(w http.ResponseWriter, name string, app, body map[string]any) {
	updated := maps.Clone(app)
	if problem := h.setCustomAppFields(name, updated, body, customAPIAppKeys); problem != "" {
		testutil.WriteError(w, http.StatusUnprocessableEntity, problem)
		return
	}
	switch {
	case updated["spec_url"] == nil:
		updated["detect_required"] = false
	case updated["url"] != app["url"] || updated["spec_url"] != app["spec_url"]:
		updated["detect_required"] = true
	}
	h.customApps[name] = updated
	testutil.WriteJSON(w, http.StatusOK, updated)
}

func (h *fakeHeimdall) customAPIApp(w http.ResponseWriter, name string) (map[string]any, bool) {
	app, ok := h.customApps[name]
	if !ok || app["type"] != "api" {
		testutil.WriteError(w, http.StatusNotFound, "custom api app "+name)
		return nil, false
	}
	return app, true
}

func (h *fakeHeimdall) getAPITools(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.customAPIApp(w, r.PathValue("name")); ok {
		tools := h.apiTools[r.PathValue("name")]
		if tools == nil {
			tools = []any{}
		}
		testutil.WriteJSON(w, http.StatusOK, tools)
	}
}

// putAPITools refuses an unknown operation_id, and needs an enabled tool
// while applications use the app. The enabled tools reach the applications,
// keeping their operations, renamed ones included.
func (h *fakeHeimdall) putAPITools(w http.ResponseWriter, r *http.Request) {
	var tools []map[string]any
	if !decodeBody(w, r, &tools) {
		return
	}
	name := r.PathValue("name")
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.customAPIApp(w, name); !ok {
		return
	}
	known := map[any]any{} // operation_id -> name
	for _, tool := range h.apiTools[name] {
		if id := tool.(map[string]any)["operation_id"]; id != nil {
			known[id] = tool.(map[string]any)["name"]
		}
	}
	renames := map[any]any{}
	var listed, enabled []any
	for _, tool := range tools {
		if id := tool["operation_id"]; id != nil {
			old, ok := known[id]
			if !ok {
				testutil.WriteError(w, http.StatusBadRequest, "unknown operation_id: "+id.(string))
				return
			}
			renames[old] = tool["name"]
		}
		listed = append(listed, tool)
		if tool["enabled"] == true {
			enabled = append(enabled, map[string]any{
				"name": tool["name"], "type": tool["type"], "description": tool["description"], "operation": nil,
			})
		}
	}
	users := h.applicationsFrom("custom-" + name)
	if len(enabled) == 0 && len(users) > 0 {
		testutil.WriteError(w, http.StatusConflict, "applications use "+name+", so it needs an enabled tool")
		return
	}
	h.apiTools[name] = listed
	for _, id := range users {
		operations := map[any]any{}
		for _, tool := range cloneTools(h.apps[id]["tools"]) {
			name := tool.(map[string]any)["name"]
			if renamed, ok := renames[name]; ok {
				name = renamed
			}
			operations[name] = tool.(map[string]any)["operation"]
		}
		merged := cloneTools(enabled)
		for _, tool := range merged {
			tool.(map[string]any)["operation"] = operations[tool.(map[string]any)["name"]]
		}
		h.apps[id]["tools"] = merged
	}
	testutil.WriteJSON(w, http.StatusOK, tools)
}

// seedDetectedAPITools stands in for a detect, which writes tools that carry
// the spec's operationId.
func seedDetectedAPITools(name string, tools ...map[string]any) {
	heimdall.mu.Lock()
	defer heimdall.mu.Unlock()
	listed := make([]any, 0, len(tools))
	for _, tool := range tools {
		listed = append(listed, tool)
	}
	heimdall.apiTools[name] = listed
}
