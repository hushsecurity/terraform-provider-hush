package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const deploymentsEndpoint = "/v1/deployments"

type Deployment struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	EnvType     string `json:"env_type"`
	Status      string `json:"status,omitempty"`
	Kind        string `json:"kind,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	ModifiedAt  string `json:"modified_at,omitempty"`
}

// CreateDeploymentInput represents the input for creating a deployment
type CreateDeploymentInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	EnvType     string `json:"env_type"`
	Kind        string `json:"kind,omitempty"`
}

// UpdateDeploymentInput represents the input for updating a deployment
type UpdateDeploymentInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	EnvType     *string `json:"env_type,omitempty"`
	Kind        *string `json:"kind,omitempty"`
}

// DeploymentCredentialsResponse embeds Deployment and adds credentials
type DeploymentCredentialsResponse struct {
	Deployment
	Token           string `json:"token"`
	Password        string `json:"password"`
	ImagePullSecret string `json:"image_pull_secret"`
}

func CreateDeployment(ctx context.Context, c *Client, input *CreateDeploymentInput) (*Deployment, error) {
	var resp DeploymentCredentialsResponse
	if err := c.doRequest(ctx, http.MethodPost, deploymentsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Deployment, nil
}

// CreateDeploymentWithCredentials creates a deployment and returns the full credentials response
func CreateDeploymentWithCredentials(ctx context.Context, c *Client, input *CreateDeploymentInput) (*DeploymentCredentialsResponse, error) {
	var resp DeploymentCredentialsResponse
	if err := c.doRequest(ctx, http.MethodPost, deploymentsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetDeployment(ctx context.Context, c *Client, id string) (*Deployment, error) {
	path := fmt.Sprintf("%s/%s", deploymentsEndpoint, id)
	var dep Deployment
	err := c.doRequest(ctx, http.MethodGet, path, nil, &dep)
	if err != nil {
		if err.Error() == "unexpected status: 404" {
			return nil, nil
		}
		return nil, err
	}
	return &dep, nil
}

func UpdateDeployment(ctx context.Context, c *Client, id string, input *UpdateDeploymentInput) (*Deployment, error) {
	path := fmt.Sprintf("%s/%s", deploymentsEndpoint, id)

	var result Deployment
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteDeployment(ctx context.Context, c *Client, id string) error {
	tflog.Info(ctx, "Deleting deployment", map[string]interface{}{
		"deployment_id": id,
	})

	path := fmt.Sprintf("%s/%s", deploymentsEndpoint, id)

	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		tflog.Error(ctx, "Failed to delete deployment", map[string]interface{}{
			"deployment_id": id,
			"error":         err.Error(),
		})
		return err
	}

	tflog.Info(ctx, "Successfully deleted deployment", map[string]interface{}{
		"deployment_id": id,
	})

	return nil
}
