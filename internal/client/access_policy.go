package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const accessPoliciesEndpoint = "/v1/access_policies"

type AttestationCriterionType string

const (
	AttestationCriterionTypeK8sNamespace      AttestationCriterionType = "k8s:ns"
	AttestationCriterionTypeK8sServiceAccount AttestationCriterionType = "k8s:sa"
	AttestationCriterionTypeK8sPodLabel       AttestationCriterionType = "k8s:pod-label"
	AttestationCriterionTypeK8sPodName        AttestationCriterionType = "k8s:pod-name"
	AttestationCriterionTypeK8sContainerName  AttestationCriterionType = "k8s:container-name"
)

type AttestationCriterion struct {
	Type  AttestationCriterionType `json:"type"`
	Key   string                   `json:"key,omitempty"`
	Value string                   `json:"value"`
}

type DeliveryType string

const (
	DeliveryTypeEnv DeliveryType = "env"
)

type DeliveryItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type DeliveryConfig struct {
	Type  DeliveryType   `json:"type"`
	Items []DeliveryItem `json:"items"`
}

type AccessPolicy struct {
	ID                  string                 `json:"id,omitempty"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	Enabled             bool                   `json:"enabled"`
	AccessCredentialID  string                 `json:"access_credential_id"`
	AttestationCriteria []AttestationCriterion `json:"attestation_criteria"`
	DeploymentIDs       []string               `json:"deployment_ids"`
	DeliveryConfig      DeliveryConfig         `json:"delivery_config"`
}

type CreateAccessPolicyInput struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	Enabled             bool                   `json:"enabled"`
	AccessCredentialID  string                 `json:"access_credential_id"`
	AttestationCriteria []AttestationCriterion `json:"attestation_criteria"`
	DeploymentIDs       []string               `json:"deployment_ids"`
	DeliveryConfig      DeliveryConfig         `json:"delivery_config"`
}

type UpdateAccessPolicyInput struct {
	Name                *string                 `json:"name,omitempty"`
	Description         *string                 `json:"description,omitempty"`
	Enabled             *bool                   `json:"enabled,omitempty"`
	AccessCredentialID  *string                 `json:"access_credential_id,omitempty"`
	AttestationCriteria *[]AttestationCriterion `json:"attestation_criteria,omitempty"`
	DeploymentIDs       *[]string               `json:"deployment_ids,omitempty"`
	DeliveryConfig      *DeliveryConfig         `json:"delivery_config,omitempty"`
}

type AccessPolicyListResponse struct {
	Items      []AccessPolicy `json:"items"`
	Total      int            `json:"total"`
	HasNext    bool           `json:"has_next"`
	NextCursor *string        `json:"next_cursor"`
}

func CreateAccessPolicy(ctx context.Context, c *Client, input *CreateAccessPolicyInput) (*AccessPolicy, error) {
	var result AccessPolicy
	if err := c.doRequest(ctx, http.MethodPost, accessPoliciesEndpoint, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func GetAccessPolicy(ctx context.Context, c *Client, id string) (*AccessPolicy, error) {
	path := fmt.Sprintf("%s/%s", accessPoliciesEndpoint, id)
	var policy AccessPolicy
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func UpdateAccessPolicy(ctx context.Context, c *Client, id string, input *UpdateAccessPolicyInput) (*AccessPolicy, error) {
	path := fmt.Sprintf("%s/%s", accessPoliciesEndpoint, id)
	var result AccessPolicy
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteAccessPolicy(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", accessPoliciesEndpoint, id)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	return nil
}

func ListAccessPolicies(ctx context.Context, c *Client) (*AccessPolicyListResponse, error) {
	params := url.Values{}
	path := accessPoliciesEndpoint
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp AccessPolicyListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
