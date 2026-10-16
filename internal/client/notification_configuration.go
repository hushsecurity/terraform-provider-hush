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
	AggregationDurationShort  AggregationDuration = "short"
	AggregationDurationDay    AggregationDuration = "day"
	AggregationDurationWeek   AggregationDuration = "week"
	AggregationDurationMonth  AggregationDuration = "month"
	AggregationDurationStream AggregationDuration = "stream"
)

type Trigger string

const (
	TriggerNewNHISecurityIssue     Trigger = "new_nhi_security_issue"
	TriggerNewAIAgentSecurityIssue Trigger = "new_ai_agent_security_issue"
	TriggerNHIDigest               Trigger = "nhi_digest"
	TriggerCertificateExpiring     Trigger = "certificate_expiring"
	TriggerAGWActivity             Trigger = "agw_activity"
)

// DispatchFilter narrows what a configuration sends. Values are numbers for
// expires_in_days and strings otherwise.
type DispatchFilter struct {
	Name   string `json:"name"`
	Op     string `json:"op"`
	Values []any  `json:"values"`
}

type NotificationConfiguration struct {
	ID              string              `json:"id,omitempty"`
	OrgID           string              `json:"org_id,omitempty"`
	Name            string              `json:"name"`
	Description     string              `json:"description,omitempty"`
	Enabled         bool                `json:"enabled"`
	ChannelIDs      []string            `json:"channel_ids"`
	Aggregation     AggregationDuration `json:"aggregation"`
	Trigger         Trigger             `json:"trigger"`
	DispatchFilters []DispatchFilter    `json:"dispatch_filters"`
	Status          string              `json:"status"`
	StatusMessage   *string             `json:"status_message"`
	LastTriggeredAt *string             `json:"last_triggered_at,omitempty"`
}

type CreateNotificationConfigurationInput struct {
	Name            string               `json:"name"`
	Description     *string              `json:"description,omitempty"`
	Enabled         bool                 `json:"enabled"`
	ChannelIDs      []string             `json:"channel_ids"`
	Aggregation     *AggregationDuration `json:"aggregation,omitempty"`
	Trigger         Trigger              `json:"trigger"`
	DispatchFilters []DispatchFilter     `json:"dispatch_filters"`
}

type UpdateNotificationConfigurationInput struct {
	Name            *string              `json:"name,omitempty"`
	Description     *string              `json:"description,omitempty"`
	Enabled         *bool                `json:"enabled,omitempty"`
	ChannelIDs      *[]string            `json:"channel_ids,omitempty"`
	Aggregation     *AggregationDuration `json:"aggregation,omitempty"`
	DispatchFilters *[]DispatchFilter    `json:"dispatch_filters,omitempty"`
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
	Items    []NotificationConfiguration `json:"items"`
	NextPage *string                     `json:"next_page"`
}

func GetNotificationConfigurationsByName(ctx context.Context, c *Client, name string) ([]NotificationConfiguration, error) {
	base := fmt.Sprintf("%s?name=%s", notificationConfigurationsEndpoint, url.QueryEscape(name))
	return collectPages(func(cursor string) ([]NotificationConfiguration, *string, error) {
		var resp NotificationConfigurationListResponse
		if err := c.doRequest(ctx, http.MethodGet, withCursor(base, cursor), nil, &resp); err != nil {
			return nil, nil, err
		}
		return resp.Items, resp.NextPage, nil
	})
}

func ListNotificationConfigurations(ctx context.Context, c *Client, enabled *bool, trigger *Trigger, aggregation *AggregationDuration) ([]NotificationConfiguration, error) {
	base := notificationConfigurationsEndpoint
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
		base = fmt.Sprintf("%s?%s", base, params.Encode())
	}

	return collectPages(func(cursor string) ([]NotificationConfiguration, *string, error) {
		var resp NotificationConfigurationListResponse
		if err := c.doRequest(ctx, http.MethodGet, withCursor(base, cursor), nil, &resp); err != nil {
			return nil, nil, err
		}
		return resp.Items, resp.NextPage, nil
	})
}

func GetNotificationConfigurationsByTrigger(ctx context.Context, c *Client, trigger string) ([]NotificationConfiguration, error) {
	triggerType := Trigger(trigger)
	return ListNotificationConfigurations(ctx, c, nil, &triggerType, nil)
}
