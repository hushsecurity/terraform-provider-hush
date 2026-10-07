package acc_tests

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"sync"

	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// fakeAgents stands in for heimdall's service agent endpoints. Like heimdall,
// it derives a typed credential's issuer and subject, binds an identity once,
// allows ten identities and two client secrets per agent, and returns a secret
// only when it is created. It also holds a user's agent, which the service
// agent resources must not take for one of theirs.
type fakeAgents struct {
	mu          sync.Mutex
	agents      map[string]map[string]any // id -> agent, without credentials
	credentials map[string][]map[string]any
	next        int
}

const userAgentID = "agt-ususer1"

var agentsFake = &fakeAgents{
	agents: map[string]map[string]any{
		userAgentID: {
			"id": userAgentID, "subject_type": "user", "name": "alice@example.com",
			"enabled": true, "acts_for_users": false,
			"resource": "https://api.example.com/v1/agents/" + userAgentID,
		},
	},
	credentials: map[string][]map[string]any{},
}

func init() {
	registerMockSetup(agentsFake.register)
}

func (f *fakeAgents) register(ms *testutil.MockServer) {
	ms.Handle("POST /v1/agents", f.create)
	ms.Handle("GET /v1/agents", f.list)
	ms.Handle("GET /v1/agents/{id}", f.get)
	ms.Handle("PATCH /v1/agents/{id}", f.patch)
	ms.Handle("DELETE /v1/agents/{id}", f.delete)
	ms.Handle("POST /v1/agents/{id}/credentials", f.addCredential)
	ms.Handle("DELETE /v1/agents/{id}/credentials/{cred}", f.removeCredential)
}

func (f *fakeAgents) out(id string) map[string]any {
	agent := maps.Clone(f.agents[id])
	credentials := []map[string]any{}
	for _, c := range f.credentials[id] {
		c = maps.Clone(c)
		delete(c, "secret")
		credentials = append(credentials, c)
	}
	agent["credentials"] = credentials
	return agent
}

func (f *fakeAgents) create(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	id := "agt-usfake" + strconv.Itoa(f.next)
	agent := map[string]any{
		"id": id, "subject_type": "service", "name": body["name"],
		"enabled": true, "acts_for_users": false,
		"resource": "https://api.example.com/v1/agents/" + id,
	}
	for _, key := range []string{"enabled", "acts_for_users"} {
		if v, ok := body[key]; ok {
			agent[key] = v
		}
	}
	f.agents[id] = agent
	testutil.WriteJSON(w, http.StatusCreated, f.out(id))
}

func (f *fakeAgents) list(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	subjectType := r.URL.Query().Get("subject_type")
	items := []map[string]any{}
	for _, id := range slices.Sorted(maps.Keys(f.agents)) {
		if subjectType == "" || f.agents[id]["subject_type"] == subjectType {
			items = append(items, f.out(id))
		}
	}
	testutil.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "next_page": nil})
}

func (f *fakeAgents) found(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if _, ok := f.agents[id]; !ok {
		testutil.WriteError(w, http.StatusNotFound, id+" not found")
		return "", false
	}
	return id, true
}

func (f *fakeAgents) get(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id, ok := f.found(w, r); ok {
		testutil.WriteJSON(w, http.StatusOK, f.out(id))
	}
}

func (f *fakeAgents) patch(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.found(w, r)
	if !ok {
		return
	}
	for key, value := range body {
		if key != "name" && key != "enabled" && key != "acts_for_users" {
			testutil.WriteError(w, http.StatusUnprocessableEntity, "unknown field "+key)
			return
		}
		f.agents[id][key] = value
	}
	testutil.WriteJSON(w, http.StatusOK, f.out(id))
}

func (f *fakeAgents) delete(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.found(w, r)
	if !ok {
		return
	}
	out := f.out(id)
	delete(f.agents, id)
	delete(f.credentials, id)
	testutil.WriteJSON(w, http.StatusOK, out)
}

// derive sets the issuer and subject heimdall derives from a typed credential.
func derive(c map[string]any) {
	switch c["type"] {
	case "aws_iam_role":
		c["subject"] = c["role_arn"]
	case "kubernetes":
		c["subject"] = fmt.Sprintf("system:serviceaccount:%v:%v", c["namespace"], c["service_account"])
	case "azure":
		c["issuer"] = fmt.Sprintf("https://login.microsoftonline.com/%v/v2.0", c["tenant_id"])
		c["subject"] = c["service_principal_id"]
	}
	if _, ok := c["conditions"]; !ok && c["type"] != "client_secret" {
		c["conditions"] = []any{}
	}
}

func (f *fakeAgents) addCredential(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if !decodeBody(w, r, &body) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.found(w, r)
	if !ok {
		return
	}
	f.next++
	credential := maps.Clone(body)
	credential["id"] = "cred" + strconv.Itoa(f.next)
	credential["created_at"] = "2026-10-07T00:00:00Z"
	if body["type"] == "client_secret" {
		secrets := 0
		for _, c := range f.credentials[id] {
			if c["type"] == "client_secret" {
				secrets++
			}
		}
		if secrets >= 2 {
			testutil.WriteError(w, http.StatusConflict, "at most 2 client secrets")
			return
		}
		secret := "hush_as_fakesecret" + strconv.Itoa(f.next)
		credential["secret"] = secret
		credential["hint"] = secret[len(secret)-4:]
		f.credentials[id] = append(f.credentials[id], credential)
		testutil.WriteJSON(w, http.StatusCreated, credential)
		return
	}
	derive(credential)
	identities := 0
	for _, c := range f.credentials[id] {
		if c["type"] != "client_secret" {
			identities++
		}
	}
	if identities >= 10 {
		testutil.WriteError(w, http.StatusConflict, "at most 10 credentials")
		return
	}
	for _, others := range f.credentials {
		for _, c := range others {
			if c["issuer"] == credential["issuer"] && c["subject"] == credential["subject"] {
				testutil.WriteError(w, http.StatusConflict, "identity already bound")
				return
			}
		}
	}
	f.credentials[id] = append(f.credentials[id], credential)
	testutil.WriteJSON(w, http.StatusCreated, credential)
}

func (f *fakeAgents) removeCredential(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.found(w, r)
	if !ok {
		return
	}
	credID := r.PathValue("cred")
	for i, c := range f.credentials[id] {
		if c["id"] == credID {
			f.credentials[id] = slices.Delete(f.credentials[id], i, i+1)
			delete(c, "secret")
			testutil.WriteJSON(w, http.StatusOK, c)
			return
		}
	}
	testutil.WriteError(w, http.StatusNotFound, credID+" not found")
}

// credentialTypes lists an agent's credential types, sorted, for a check.
func (f *fakeAgents) credentialTypes(name string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var types []string
	for id, agent := range f.agents {
		if agent["name"] != name {
			continue
		}
		for _, c := range f.credentials[id] {
			types = append(types, c["type"].(string))
		}
	}
	slices.Sort(types)
	return types
}
