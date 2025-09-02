package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const notificationConfigurationsEndpoint = "/v1/notification_configurations"

// AggregationDuration represents the aggregation duration for notifications
type AggregationDuration string

const (
	AggregationDurationShort AggregationDuration = "short"
	AggregationDurationWeek  AggregationDuration = "week"
	AggregationDurationMonth AggregationDuration = "month"
)

// Trigger represents the notification trigger type
type Trigger string

const (
	TriggerNewNHIAtRisk Trigger = "new_nhi_at_risk"
	TriggerNHIDigest    Trigger = "nhi_digest"
)

// NotificationConfiguration represents a notification configuration
type NotificationConfiguration struct {
	ID              string              `json:"id,omitempty"`
	OrgID           string              `json:"org_id,omitempty"`
	Name            string              `json:"name"`
	Description     string              `json:"description,omitempty"`
	Enabled         bool                `json:"enabled"`
	ChannelIDs      []string            `json:"channel_ids"`
	Aggregation     AggregationDuration `json:"aggregation"`
	Trigger         Trigger             `json:"trigger"`
	CreatedAt       string              `json:"created_at,omitempty"`
	ModifiedAt      string              `json:"modified_at,omitempty"`
	LastTriggeredAt *string             `json:"last_triggered_at,omitempty"`
}

// CreateNotificationConfigurationInput represents the input for creating a notification configuration
type CreateNotificationConfigurationInput struct {
	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	Enabled     bool                `json:"enabled"`
	ChannelIDs  []string            `json:"channel_ids"`
	Aggregation AggregationDuration `json:"aggregation"`
	Trigger     Trigger             `json:"trigger"`
}

// UpdateNotificationConfigurationInput represents the input for updating a notification configuration
type UpdateNotificationConfigurationInput struct {
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Enabled     *bool                `json:"enabled,omitempty"`
	ChannelIDs  *[]string            `json:"channel_ids,omitempty"`
	Aggregation *AggregationDuration `json:"aggregation,omitempty"`
	Trigger     *Trigger             `json:"trigger,omitempty"`
}

// CreateNotificationConfiguration creates a new notification configuration
func CreateNotificationConfiguration(ctx context.Context, c *Client, input *CreateNotificationConfigurationInput) (*NotificationConfiguration, error) {
	var resp NotificationConfiguration
	if err := c.doRequest(ctx, http.MethodPost, notificationConfigurationsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetNotificationConfiguration retrieves a notification configuration by ID
func GetNotificationConfiguration(ctx context.Context, c *Client, id string) (*NotificationConfiguration, error) {
	path := fmt.Sprintf("%s/%s", notificationConfigurationsEndpoint, id)
	var config NotificationConfiguration
	err := c.doRequest(ctx, http.MethodGet, path, nil, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateNotificationConfiguration updates a notification configuration
func UpdateNotificationConfiguration(ctx context.Context, c *Client, id string, input *UpdateNotificationConfigurationInput) (*NotificationConfiguration, error) {
	path := fmt.Sprintf("%s/%s", notificationConfigurationsEndpoint, id)
	var result NotificationConfiguration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteNotificationConfiguration deletes a notification configuration
func DeleteNotificationConfiguration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", notificationConfigurationsEndpoint, id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// NotificationConfigurationListResponse represents the response from listing notification configurations
type NotificationConfigurationListResponse struct {
	Items      []NotificationConfiguration `json:"items"`
	NextCursor *string                     `json:"next_cursor"`
	HasMore    bool                        `json:"has_more"`
}

// GetNotificationConfigurationsByName retrieves notification configurations by name
func GetNotificationConfigurationsByName(ctx context.Context, c *Client, name string) ([]NotificationConfiguration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s", notificationConfigurationsEndpoint, encodedName)

	var resp NotificationConfigurationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// ListNotificationConfigurations retrieves all notification configurations with optional filters
func ListNotificationConfigurations(ctx context.Context, c *Client, enabled *bool, trigger *Trigger, aggregation *AggregationDuration) ([]NotificationConfiguration, error) {
	path := notificationConfigurationsEndpoint
	params := url.Values{}

	if enabled != nil {
		params.Add("enabled", fmt.Sprintf("%t", *enabled))
	}
	if trigger != nil {
		params.Add("trigger", string(*trigger))
	}
	if aggregation != nil {
		params.Add("aggregation", string(*aggregation))
	}

	if len(params) > 0 {
		path = fmt.Sprintf("%s?%s", path, params.Encode())
	}

	var resp NotificationConfigurationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}
