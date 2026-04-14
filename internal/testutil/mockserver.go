package testutil

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
)

// Operation represents a CRUD operation type.
type Operation string

const (
	OpCreate Operation = "create"
	OpRead   Operation = "read"
	OpUpdate Operation = "update"
	OpDelete Operation = "delete"
)

// HookError allows hooks to override responses.
type HookError struct {
	Status int
	Detail string
}

// HookFunc can modify objects or return errors.
type HookFunc func(op Operation, obj map[string]any) *HookError

// route represents a parsed endpoint pattern.
type route struct {
	method   string
	pattern  *regexp.Regexp
	template string // original pattern like "/v1/deployments/{deployment_id}"
	response map[string]any
}

// MockServer is a dynamic mock HTTP server driven by fixtures.
type MockServer struct {
	Server   *httptest.Server
	store    map[string]map[string]any // resourceKey -> id -> object
	routes   []route
	hooks    map[string]map[Operation][]HookFunc
	fixtures *Fixtures
	mu       sync.RWMutex
}

// NewMockServer creates a mock server from fixtures.
func NewMockServer(f *Fixtures) *MockServer {
	ms := &MockServer{
		store:    make(map[string]map[string]any),
		routes:   parseRoutes(f.Endpoints),
		hooks:    make(map[string]map[Operation][]HookFunc),
		fixtures: f,
	}
	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// URL returns the mock server URL.
func (ms *MockServer) URL() string { return ms.Server.URL }

// Close shuts down the mock server.
func (ms *MockServer) Close() { ms.Server.Close() }

// SeedObject inserts a pre-existing object into the mock store.
// Useful for resources that require pre-existing data (e.g., predefined configs).
func (ms *MockServer) SeedObject(storeKey, id string, obj map[string]any) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.store[storeKey] == nil {
		ms.store[storeKey] = make(map[string]any)
	}
	ms.store[storeKey][id] = obj
}

// OnOperation registers a hook for a resource type and operation.
func (ms *MockServer) OnOperation(resourceType string, op Operation, hook HookFunc) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.hooks[resourceType] == nil {
		ms.hooks[resourceType] = make(map[Operation][]HookFunc)
	}
	ms.hooks[resourceType][op] = append(ms.hooks[resourceType][op], hook)
}

func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	// Handle OAuth token endpoint
	if r.URL.Path == "/v1/oauth/token" && r.Method == http.MethodPost {
		ms.handleAuth(w)
		return
	}

	// Match against routes
	for _, rt := range ms.routes {
		if r.Method != rt.method {
			continue
		}
		if matches := rt.pattern.FindStringSubmatch(r.URL.Path); matches != nil {
			ms.handleRoute(w, r, rt, matches)
			return
		}
	}

	ms.writeError(w, 404, "endpoint not found: "+r.Method+" "+r.URL.Path)
}

func (ms *MockServer) handleAuth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": "mock-token-" + generateUUID(),
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func (ms *MockServer) handleRoute(w http.ResponseWriter, r *http.Request, rt route, matches []string) {
	resourceKey := extractResourceKey(rt.template)
	storeKey := normalizeStoreKey(resourceKey)
	var id string
	if len(matches) > 1 {
		id = matches[1]
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.store[storeKey] == nil {
		ms.store[storeKey] = make(map[string]any)
	}

	switch r.Method {
	case http.MethodPost:
		ms.handleCreate(w, r, resourceKey, storeKey, rt.response)
	case http.MethodGet:
		if id != "" {
			ms.handleGet(w, storeKey, id, rt.response)
		} else {
			ms.handleList(w, r, storeKey)
		}
	case http.MethodPatch:
		ms.handleUpdate(w, r, resourceKey, storeKey, id, rt.response)
	case http.MethodDelete:
		ms.handleDelete(w, storeKey, id)
	default:
		ms.writeError(w, 405, "method not allowed")
	}
}

func (ms *MockServer) handleCreate(w http.ResponseWriter, r *http.Request, resourceKey, storeKey string, template map[string]any) {
	var obj map[string]any
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		ms.writeError(w, 400, "invalid JSON body")
		return
	}

	uuid := generateUUID()
	id := ms.generateID(resourceKey, uuid)
	obj["id"] = id

	// Set terminal status so provider status-polling succeeds immediately.
	// Resources that implement statusFields() poll until status is "ok".
	// Hooks can override this for testing non-happy paths.
	if _, exists := obj["status"]; !exists {
		obj["status"] = "ok"
	}
	if _, exists := obj["status_detail"]; !exists {
		obj["status_detail"] = ""
	}

	// Merge computed fields
	computed := ms.fixtures.GetComputedFields(ms.getComputedFieldsKey(resourceKey), uuid)
	for k, v := range computed {
		if _, exists := obj[k]; !exists {
			obj[k] = v
		}
	}

	// Apply hooks
	if err := ms.runHooks(resourceKey, OpCreate, obj); err != nil {
		ms.writeError(w, err.Status, err.Detail)
		return
	}

	ms.store[storeKey][id] = obj
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(obj)
}

func (ms *MockServer) handleGet(w http.ResponseWriter, storeKey, id string, template map[string]any) {
	obj, ok := ms.store[storeKey][id]
	if !ok {
		ms.writeError(w, 404, id+" not found")
		return
	}

	objMap := obj.(map[string]any)
	if err := ms.runHooks(storeKey, OpRead, objMap); err != nil {
		ms.writeError(w, err.Status, err.Detail)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(obj)
}

func (ms *MockServer) handleList(w http.ResponseWriter, r *http.Request, storeKey string) {
	var items []any
	for _, obj := range ms.store[storeKey] {
		if ms.matchesFilters(obj.(map[string]any), r.URL.Query()) {
			items = append(items, obj)
		}
	}
	if items == nil {
		items = []any{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items":    items,
		"total":    len(items),
		"has_more": false,
		"has_next": false,
	})
}

func (ms *MockServer) handleUpdate(w http.ResponseWriter, r *http.Request, resourceKey, storeKey, id string, template map[string]any) {
	obj, ok := ms.store[storeKey][id]
	if !ok {
		ms.writeError(w, 404, id+" not found")
		return
	}

	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		ms.writeError(w, 400, "invalid JSON body")
		return
	}

	objMap := obj.(map[string]any)
	for k, v := range updates {
		objMap[k] = v
	}

	if err := ms.runHooks(resourceKey, OpUpdate, objMap); err != nil {
		ms.writeError(w, err.Status, err.Detail)
		return
	}

	ms.store[storeKey][id] = objMap
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(objMap)
}

func (ms *MockServer) handleDelete(w http.ResponseWriter, storeKey, id string) {
	if _, ok := ms.store[storeKey][id]; !ok {
		ms.writeError(w, 404, id+" not found")
		return
	}

	obj := ms.store[storeKey][id].(map[string]any)
	if err := ms.runHooks(storeKey, OpDelete, obj); err != nil {
		ms.writeError(w, err.Status, err.Detail)
		return
	}

	delete(ms.store[storeKey], id)
	w.WriteHeader(204)
}

func (ms *MockServer) runHooks(resourceKey string, op Operation, obj map[string]any) *HookError {
	// Try exact match first, then base resource type
	keys := []string{resourceKey, ms.getComputedFieldsKey(resourceKey)}
	for _, key := range keys {
		if hooks, ok := ms.hooks[key]; ok {
			for _, hook := range hooks[op] {
				if err := hook(op, obj); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ms *MockServer) matchesFilters(obj map[string]any, params map[string][]string) bool {
	for key, values := range params {
		if len(values) == 0 {
			continue
		}
		val := fmt.Sprintf("%v", obj[key])
		if val != values[0] {
			return false
		}
	}
	return true
}

func (ms *MockServer) generateID(resourceKey, uuid string) string {
	prefix := ms.getIDPrefix(resourceKey)
	return prefix + "-" + uuid[:8]
}

func (ms *MockServer) getIDPrefix(resourceKey string) string {
	prefixes := map[string]string{
		"deployments":                 "dep",
		"access_credentials":          "acr",
		"access_policies":             "apl",
		"access_privileges":           "apr",
		"notification_channels":       "nch",
		"notification_configurations": "ncf",
		"integrations":                "int",
	}
	// Check for base resource type
	for base, prefix := range prefixes {
		if strings.Contains(resourceKey, base) {
			return prefix
		}
	}
	// Default: first 3 chars
	if len(resourceKey) >= 3 {
		return resourceKey[:3]
	}
	return "id"
}

func (ms *MockServer) getComputedFieldsKey(resourceKey string) string {
	// Map resource paths to computed_fields keys
	if strings.Contains(resourceKey, "access_credentials") {
		return "access_credential"
	}
	if strings.Contains(resourceKey, "access_policies") {
		return "access_policy"
	}
	if strings.Contains(resourceKey, "access_privileges") {
		return "access_privilege"
	}
	if strings.Contains(resourceKey, "deployments") {
		return "deployment"
	}
	if strings.Contains(resourceKey, "notification_channels") {
		return "notification_channel"
	}
	if strings.Contains(resourceKey, "notification_configurations") {
		return "notification_configuration"
	}
	if strings.Contains(resourceKey, "integrations") {
		return "gcp_integration"
	}
	return resourceKey
}

func (ms *MockServer) writeError(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      status,
		"status_code": status,
		"detail":      detail,
		"title":       http.StatusText(status),
	})
}

// parseRoutes converts fixture endpoints to route patterns.
func parseRoutes(endpoints map[string]map[string]any) []route {
	var routes []route
	for endpoint, responses := range endpoints {
		parts := strings.SplitN(endpoint, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method, path := parts[0], parts[1]

		// Convert path params like {id} to regex
		pattern := "^" + regexp.QuoteMeta(path) + "$"
		pattern = regexp.MustCompile(`\\{[^}]+\\}`).ReplaceAllString(pattern, "([^/]+)")

		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		var resp map[string]any
		if r200, ok := responses["200"].(map[string]any); ok {
			resp = r200
		}

		routes = append(routes, route{
			method:   method,
			pattern:  re,
			template: path,
			response: resp,
		})
	}
	return routes
}

// extractResourceKey extracts resource identifier from path template.
func extractResourceKey(template string) string {
	// Remove /v1/ prefix and parameter placeholders
	path := strings.TrimPrefix(template, "/v1/")
	// Remove trailing /{param}
	if idx := strings.LastIndex(path, "/{"); idx != -1 {
		path = path[:idx]
	}
	return path
}

// normalizeStoreKey maps subtype-specific paths to a shared store key.
// E.g. "access_credentials/postgres" and "access_credentials" share one store.
func normalizeStoreKey(resourceKey string) string {
	bases := []string{"access_credentials", "access_privileges", "access_policies"}
	for _, base := range bases {
		if strings.HasPrefix(resourceKey, base) {
			return base
		}
	}
	return resourceKey
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
