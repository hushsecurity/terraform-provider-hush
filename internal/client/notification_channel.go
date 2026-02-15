package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const notificationChannelsEndpoint = "/v1/notification_channels"

type NotificationChannelType string

const (
	NotificationChannelTypeEmail   NotificationChannelType = "email"
	NotificationChannelTypeWebhook NotificationChannelType = "webhook"
	NotificationChannelTypeSlack   NotificationChannelType = "slack"
)

type ConfigBase interface {
	GetType() NotificationChannelType
}

type EmailConfig struct {
	Addresses []string `json:"addresses,omitempty"`
	Address   string   `json:"address,omitempty"`
	Verified  bool     `json:"verified,omitempty"`
}

func (e EmailConfig) GetType() NotificationChannelType {
	return NotificationChannelTypeEmail
}

type WebhookMethod string

const (
	WebhookMethodPOST WebhookMethod = "POST"
	WebhookMethodGET  WebhookMethod = "GET"
)

type WebhookConfig struct {
	URL      string        `json:"url"`
	Method   WebhookMethod `json:"method"`
	Verified bool          `json:"verified,omitempty"`
}

func (w WebhookConfig) GetType() NotificationChannelType {
	return NotificationChannelTypeWebhook
}

type SlackConfig struct {
	IntegrationID string `json:"integration_id"`
	Channel       string `json:"channel"`
	ChannelID     string `json:"channel_id"`
}

func (s SlackConfig) GetType() NotificationChannelType {
	return NotificationChannelTypeSlack
}

type NotificationChannel struct {
	ID          string                   `json:"id,omitempty"`
	OrgID       string                   `json:"org_id,omitempty"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Enabled     bool                     `json:"enabled"`
	Type        NotificationChannelType  `json:"type"`
	Config      []map[string]interface{} `json:"config"`
}

type CreateNotificationChannelInput struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	Enabled     bool                     `json:"enabled"`
	Config      []map[string]interface{} `json:"config"`
}

type UpdateNotificationChannelInput struct {
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Enabled     *bool                     `json:"enabled,omitempty"`
	Type        *NotificationChannelType  `json:"type,omitempty"`
	Config      *[]map[string]interface{} `json:"config,omitempty"`
}

func CreateNotificationChannel(ctx context.Context, c *Client, input *CreateNotificationChannelInput) (*NotificationChannel, error) {
	var resp NotificationChannel
	if err := c.doRequest(ctx, http.MethodPost, notificationChannelsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetNotificationChannel(ctx context.Context, c *Client, id string) (*NotificationChannel, error) {
	path := fmt.Sprintf("%s/%s", notificationChannelsEndpoint, id)
	var channel NotificationChannel
	err := c.doRequest(ctx, http.MethodGet, path, nil, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func UpdateNotificationChannel(ctx context.Context, c *Client, id string, input *UpdateNotificationChannelInput) (*NotificationChannel, error) {
	path := fmt.Sprintf("%s/%s", notificationChannelsEndpoint, id)
	var result NotificationChannel
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteNotificationChannel(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", notificationChannelsEndpoint, id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

type NotificationChannelListResponse struct {
	Items      []NotificationChannel `json:"items"`
	NextCursor *string               `json:"next_cursor"`
	HasMore    bool                  `json:"has_more"`
}

func GetNotificationChannelsByName(ctx context.Context, c *Client, name string) ([]NotificationChannel, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s", notificationChannelsEndpoint, encodedName)

	var resp NotificationChannelListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func ListNotificationChannels(ctx context.Context, c *Client, enabled *bool) ([]NotificationChannel, error) {
	path := notificationChannelsEndpoint
	if enabled != nil {
		path = fmt.Sprintf("%s?enabled=%t", path, *enabled)
	}

	var resp NotificationChannelListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}
