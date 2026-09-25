package acc_tests

import (
	"encoding/json"
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
}

var heimdall = &fakeHeimdall{
	consentMethods: map[string][]string{},
}

func init() {
	registerMockSetup(heimdall.register)
}

func (h *fakeHeimdall) register(ms *testutil.MockServer) {
	ms.Handle("GET /v1/deployments/{deployment_id}/agw/consent_methods", h.getConsentMethods)
	ms.Handle("PUT /v1/deployments/{deployment_id}/agw/consent_methods", h.putConsentMethods)
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
