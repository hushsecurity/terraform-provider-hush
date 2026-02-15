package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const notificationConfigurationsEndpoint = "/v1/notification_configurations"

type AggregationDuration string

const (
	AggregationDurationShort AggregationDuration = "short"
	AggregationDurationWeek  AggregationDuration = "week"
	AggregationDurationMonth AggregationDuration = "month"
)

type Trigger string

const (
	TriggerNewNHIAtRisk Trigger = "new_nhi_at_risk"
	TriggerNHIDigest    Trigger = "nhi_digest"
)

type NotificationConfiguration struct {
	ID              string              `json:"id,omitempty"`
	OrgID           string              `json:"org_id,omitempty"`
	Name            string              `json:"name"`
	Description     string              `json:"description,omitempty"`
	Enabled         bool                `json:"enabled"`
	ChannelIDs      []string            `json:"channel_ids"`
	Aggregation     AggregationDuration `json:"aggregation"`
	Trigger         Trigger             `json:"trigger"`
	LastTriggeredAt *string             `json:"last_triggered_at,omitempty"`
}

type CreateNotificationConfigurationInput struct {
	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	Enabled     bool                `json:"enabled"`
	ChannelIDs  []string            `json:"channel_ids"`
	Aggregation AggregationDuration `json:"aggregation"`
	Trigger     Trigger             `json:"trigger"`
}

type UpdateNotificationConfigurationInput struct {
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Enabled     *bool                `json:"enabled,omitempty"`
	ChannelIDs  *[]string            `json:"channel_ids,omitempty"`
	Aggregation *AggregationDuration `json:"aggregation,omitempty"`
	Trigger     *Trigger             `json:"trigger,omitempty"`
}

func CreateNotificationConfiguration(ctx context.Context, c *Client, input *CreateNotificationConfigurationInput) (*NotificationConfiguration, error) {
	var resp NotificationConfiguration
	if err := c.doRequest(ctx, http.MethodPost, notificationConfigurationsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetNotificationConfiguration(ctx context.Context, c *Client, id string) (*NotificationConfiguration, error) {
	path := fmt.Sprintf("%s/%s", notificationConfigurationsEndpoint, id)
	var config NotificationConfiguration
	err := c.doRequest(ctx, http.MethodGet, path, nil, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func UpdateNotificationConfiguration(ctx context.Context, c *Client, id string, input *UpdateNotificationConfigurationInput) (*NotificationConfiguration, error) {
	path := fmt.Sprintf("%s/%s", notificationConfigurationsEndpoint, id)
	var result NotificationConfiguration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteNotificationConfiguration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", notificationConfigurationsEndpoint, id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

type NotificationConfigurationListResponse struct {
	Items      []NotificationConfiguration `json:"items"`
	NextCursor *string                     `json:"next_cursor"`
	HasMore    bool                        `json:"has_more"`
}

func GetNotificationConfigurationsByName(ctx context.Context, c *Client, name string) ([]NotificationConfiguration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s", notificationConfigurationsEndpoint, encodedName)

	var resp NotificationConfigurationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

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

func GetNotificationConfigurationsByTrigger(ctx context.Context, c *Client, trigger string) ([]NotificationConfiguration, error) {
	triggerType := Trigger(trigger)
	return ListNotificationConfigurations(ctx, c, nil, &triggerType, nil)
}
