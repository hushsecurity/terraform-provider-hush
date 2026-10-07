package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// The credential types a service agent signs in with. client_secret is
// managed on its own (hush_service_agent_client_secret); the others are the
// agent's federated identities.
const (
	CredentialAWSIAMRole   = "aws_iam_role"
	CredentialKubernetes   = "kubernetes"
	CredentialAzure        = "azure"
	CredentialOIDC         = "oidc"
	CredentialClientSecret = "client_secret"
)

// FederatedCredentialTypes are the types managed inline on hush_service_agent.
var FederatedCredentialTypes = []string{
	CredentialAWSIAMRole, CredentialKubernetes, CredentialAzure, CredentialOIDC,
}

// heimdall's limits per agent.
const (
	MaxAgentCredentials   = 10
	MaxAgentClientSecrets = 2
)

// ClaimOps are the comparisons a claim condition can make.
var ClaimOps = []string{"eq", "pfx", "sfx"}

// ClaimCondition is a check on a verified claim. Claim is a path into the
// claims; the provider always sends a list, which heimdall returns as is.
type ClaimCondition struct {
	Claim []string `json:"claim"`
	Op    string   `json:"op"`
	Value string   `json:"value"`
}

// UnmarshalJSON accepts a claim given as one name, as heimdall also allows.
func (cc *ClaimCondition) UnmarshalJSON(data []byte) error {
	var raw struct {
		Claim json.RawMessage `json:"claim"`
		Op    string          `json:"op"`
		Value string          `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	cc.Op, cc.Value, cc.Claim = raw.Op, raw.Value, nil
	if len(raw.Claim) == 0 || string(raw.Claim) == "null" {
		return nil
	}
	var one string
	if err := json.Unmarshal(raw.Claim, &one); err == nil {
		cc.Claim = []string{one}
		return nil
	}
	return json.Unmarshal(raw.Claim, &cc.Claim)
}

// AgentCredential is one credential as heimdall returns it. Which fields are
// set depends on Type; Issuer and Subject are set on every federated type,
// derived by heimdall from the typed fields where it can.
type AgentCredential struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	Issuer     string           `json:"issuer,omitempty"`
	Subject    string           `json:"subject,omitempty"`
	Conditions []ClaimCondition `json:"conditions,omitempty"`
	// aws_iam_role
	RoleARN string `json:"role_arn,omitempty"`
	// kubernetes
	Namespace      string `json:"namespace,omitempty"`
	ServiceAccount string `json:"service_account,omitempty"`
	// azure
	TenantID           string `json:"tenant_id,omitempty"`
	ServicePrincipalID string `json:"service_principal_id,omitempty"`
	Audience           string `json:"audience,omitempty"`
	// client_secret
	Hint   string `json:"hint,omitempty"`
	Secret string `json:"secret,omitempty"` // only in the response that creates it
}

type ServiceAgent struct {
	ID           string            `json:"id"`
	SubjectType  string            `json:"subject_type"`
	Name         string            `json:"name"`
	Enabled      bool              `json:"enabled"`
	ActsForUsers bool              `json:"acts_for_users"`
	Resource     string            `json:"resource"`
	Credentials  []AgentCredential `json:"credentials"`
}

type ServiceAgentInput struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	ActsForUsers bool   `json:"acts_for_users"`
}

// ServiceAgentUpdate is a patch: a nil field is left as it is.
type ServiceAgentUpdate struct {
	Name         *string `json:"name,omitempty"`
	Enabled      *bool   `json:"enabled,omitempty"`
	ActsForUsers *bool   `json:"acts_for_users,omitempty"`
}

const agentsEndpoint = "/v1/agents"

func CreateServiceAgent(ctx context.Context, c *Client, input *ServiceAgentInput) (*ServiceAgent, error) {
	var agent ServiceAgent
	if err := c.doRequest(ctx, http.MethodPost, agentsEndpoint, input, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// ErrNotServiceAgent is returned for an agent id that names a person's agent
// (one they signed in through), which the same endpoint also serves.
var ErrNotServiceAgent = errors.New("not a service agent")

func GetServiceAgent(ctx context.Context, c *Client, id string) (*ServiceAgent, error) {
	var agent ServiceAgent
	if err := c.doRequest(ctx, http.MethodGet, agentsEndpoint+"/"+url.PathEscape(id), nil, &agent); err != nil {
		return nil, err
	}
	if agent.SubjectType != "service" {
		return nil, fmt.Errorf("agent %s: %w", id, ErrNotServiceAgent)
	}
	return &agent, nil
}

// ListServiceAgents returns every service agent in the organization.
func ListServiceAgents(ctx context.Context, c *Client) ([]ServiceAgent, error) {
	path := agentsEndpoint + "?subject_type=service"
	return collectPages(func(cursor string) ([]ServiceAgent, *string, error) {
		var page struct {
			Items    []ServiceAgent `json:"items"`
			NextPage *string        `json:"next_page"`
		}
		if err := c.doRequest(ctx, http.MethodGet, withCursor(path, cursor), nil, &page); err != nil {
			return nil, nil, err
		}
		return page.Items, page.NextPage, nil
	})
}

func UpdateServiceAgent(ctx context.Context, c *Client, id string, input *ServiceAgentUpdate) (*ServiceAgent, error) {
	var agent ServiceAgent
	if err := c.doRequest(ctx, http.MethodPatch, agentsEndpoint+"/"+url.PathEscape(id), input, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// DeleteServiceAgent also removes its credentials and the users it acted for.
func DeleteServiceAgent(ctx context.Context, c *Client, id string) error {
	return c.doRequest(ctx, http.MethodDelete, agentsEndpoint+"/"+url.PathEscape(id), nil, nil)
}

// AddAgentCredential binds a credential to the agent. Credentials can't be
// changed, only added and removed. For a client secret, pass
// {"type": "client_secret"}: the response carries the secret, the only time
// it is returned.
func AddAgentCredential(ctx context.Context, c *Client, agentID string, credential map[string]any) (*AgentCredential, error) {
	var out AgentCredential
	path := fmt.Sprintf("%s/%s/credentials", agentsEndpoint, url.PathEscape(agentID))
	if err := c.doRequest(ctx, http.MethodPost, path, credential, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func RemoveAgentCredential(ctx context.Context, c *Client, agentID, credentialID string) error {
	path := fmt.Sprintf("%s/%s/credentials/%s", agentsEndpoint, url.PathEscape(agentID), url.PathEscape(credentialID))
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
