package client

import (
	"context"
	"fmt"
	"net/http"
)

// The consent methods a deployment's agent gateway may prompt a user through,
// kept by heimdall per deployment.
const (
	ConsentMethodOIDCCallback   = "oidc_callback"
	ConsentMethodPushSlack      = "push_slack"
	ConsentMethodMCPElicitation = "mcp_elicitation"
)

// ConsentMethods are the methods the provider accepts. mcp_elicitation is
// among them because heimdall accepts it, and refusing it would take a
// provider release once it works; the gateway does not act on it yet, so the
// documentation and examples leave it out on purpose.
var ConsentMethods = []string{
	ConsentMethodOIDCCallback,
	ConsentMethodPushSlack,
	ConsentMethodMCPElicitation,
}

// ConsentMethod is one entry of what the API returns: the method and a
// description of how it prompts the user, which the API writes itself.
type ConsentMethod struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func agwConsentMethodsPath(deploymentID string) string {
	return fmt.Sprintf("/v1/deployments/%s/agw/consent_methods", deploymentID)
}

// GetAgwConsentMethods returns the methods stored for a deployment. A
// deployment nothing was ever stored for answers 404, as does one that does
// not exist.
func GetAgwConsentMethods(ctx context.Context, c *Client, deploymentID string) ([]ConsentMethod, error) {
	var methods []ConsentMethod
	if err := c.doRequest(ctx, http.MethodGet, agwConsentMethodsPath(deploymentID), nil, &methods); err != nil {
		return nil, err
	}
	return methods, nil
}

// SetAgwConsentMethods replaces the methods of a deployment. The body is the
// bare list of names: the API takes no object around it.
func SetAgwConsentMethods(ctx context.Context, c *Client, deploymentID string, methods []string) ([]ConsentMethod, error) {
	var result []ConsentMethod
	if err := c.doRequest(ctx, http.MethodPut, agwConsentMethodsPath(deploymentID), methods, &result); err != nil {
		return nil, err
	}
	return result, nil
}
