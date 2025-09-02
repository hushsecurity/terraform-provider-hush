package notification_channel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the notification channel"
	nameDesc        = "The name of the notification channel"
	descriptionDesc = "The description of the notification channel"
	enabledDesc     = "Whether the notification channel is enabled"
	createdAtDesc   = "The creation timestamp of the notification channel"
	modifiedAtDesc  = "The last modification timestamp of the notification channel"
)

func NotificationChannelResourceSchema() map[string]*schema.Schema {
	s := NotificationChannelDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description: nameDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["description"] = &schema.Schema{
		Description: descriptionDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["enabled"] = &schema.Schema{
		Description: enabledDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
	}

	s["email_config"] = &schema.Schema{
		Description:   "Email notification configuration",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100,
		ConflictsWith: []string{"webhook_config", "slack_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address": {
					Description: "Email address for notifications",
					Type:        schema.TypeString,
					Required:    true,
				},
				"verified": {
					Description: "Whether the email address is verified",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			},
		},
	}

	s["webhook_config"] = &schema.Schema{
		Description:   "Webhook notification configuration. Multiple webhook_config blocks can be specified to send notifications to multiple webhook URLs.",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100, // API limit as confirmed in analysis
		ConflictsWith: []string{"email_config", "slack_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Description: "Webhook URL",
					Type:        schema.TypeString,
					Required:    true,
				},
				"method": {
					Description: "HTTP method for webhook requests",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "POST",
					ValidateFunc: validation.StringInSlice([]string{
						string(client.WebhookMethodPOST),
						string(client.WebhookMethodGET),
					}, false),
				},
				"verified": {
					Description: "Whether the webhook URL is verified",
					Type:        schema.TypeBool,
					Computed:    true,
				},
			},
		},
	}

	s["slack_config"] = &schema.Schema{
		Description:   "Slack notification configuration. Multiple slack_config blocks can be specified to send notifications to multiple Slack channels.",
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      100, // API limit as confirmed in analysis
		ConflictsWith: []string{"email_config", "webhook_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"integration_id": {
					Description: "Slack integration ID",
					Type:        schema.TypeString,
					Required:    true,
				},
				"channel": {
					Description: "Slack channel name",
					Type:        schema.TypeString,
					Required:    true,
				},
				"channel_id": {
					Description: "Slack channel ID",
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}

	return s
}

func NotificationChannelDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description:   idDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"id"},
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enabled": {
			Description: enabledDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"type": {
			Description: "The type of notification channel",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"email_config": {
			Description: "Email notification configuration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"address": {
						Description: "Email address for notifications",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"verified": {
						Description: "Whether the email address is verified",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
		"webhook_config": {
			Description: "Webhook notification configuration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Description: "Webhook URL",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"method": {
						Description: "HTTP method for webhook requests",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"verified": {
						Description: "Whether the webhook URL is verified",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
		"slack_config": {
			Description: "Slack notification configuration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"integration_id": {
						Description: "Slack integration ID",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"channel": {
						Description: "Slack channel name",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"channel_id": {
						Description: "Slack channel ID",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"created_at": {
			Description: createdAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"modified_at": {
			Description: modifiedAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

// Helper Functions

func notificationChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	var channel *client.NotificationChannel
	var err error

	if id := d.Id(); id != "" {
		channel, err = client.GetNotificationChannel(ctx, c, id)
		if err != nil {
			// Handle 404 errors gracefully by removing from state
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			} else {
				return diag.FromErr(err)
			}
		}
	} else if id, exists := d.GetOk("id"); exists {
		channelID := id.(string)
		channel, err = client.GetNotificationChannel(ctx, c, channelID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no notification channel found with ID: %s", channelID)
			} else {
				return diag.FromErr(err)
			}
		}
	} else if name, exists := d.GetOk("name"); exists {
		channelName := name.(string)
		channels, err := client.GetNotificationChannelsByName(ctx, c, channelName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup notification channel by name '%s': %w", channelName, err))
		}

		switch len(channels) {
		case 0:
			return diag.Errorf("no notification channel found with name: %s", channelName)
		case 1:
			channel = &channels[0]
		default:
			return diag.Errorf("multiple notification channels found with name '%s'. Channel names must be unique. Consider using the channel ID instead for exact matching", channelName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(channel.ID)
	}

	if diags := setNotificationChannelFields(d, channel); diags.HasError() {
		return diags
	}

	return nil
}

func setNotificationChannelFields(d *schema.ResourceData, channel *client.NotificationChannel) diag.Diagnostics {
	fields := map[string]interface{}{
		"name":        channel.Name,
		"description": channel.Description,
		"enabled":     channel.Enabled,
		"type":        string(channel.Type),
		"created_at":  channel.CreatedAt,
		"modified_at": channel.ModifiedAt,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setNotificationChannelConfigFields(d *schema.ResourceData, channel *client.NotificationChannel) error {
	if err := d.Set("email_config", nil); err != nil {
		return fmt.Errorf("failed to clear email_config: %w", err)
	}
	if err := d.Set("webhook_config", nil); err != nil {
		return fmt.Errorf("failed to clear webhook_config: %w", err)
	}
	if err := d.Set("slack_config", nil); err != nil {
		return fmt.Errorf("failed to clear slack_config: %w", err)
	}

	switch channel.Type {
	case client.NotificationChannelTypeEmail:
		if len(channel.Config) > 0 {
			emailConfigs := make([]map[string]interface{}, len(channel.Config))
			for i, config := range channel.Config {
				emailConfigs[i] = map[string]interface{}{
					"address":  config["address"],
					"verified": config["verified"],
				}
			}
			if err := d.Set("email_config", emailConfigs); err != nil {
				return fmt.Errorf("failed to set email_config: %w", err)
			}
		}
	case client.NotificationChannelTypeWebhook:
		if len(channel.Config) > 0 {
			webhookConfigs := make([]map[string]interface{}, len(channel.Config))
			for i, config := range channel.Config {
				webhookConfigs[i] = map[string]interface{}{
					"url":      config["url"],
					"method":   config["method"],
					"verified": config["verified"],
				}
			}
			if err := d.Set("webhook_config", webhookConfigs); err != nil {
				return fmt.Errorf("failed to set webhook_config: %w", err)
			}
		}
	case client.NotificationChannelTypeSlack:
		if len(channel.Config) > 0 {
			slackConfigs := make([]map[string]interface{}, len(channel.Config))
			for i, config := range channel.Config {
				slackConfigs[i] = map[string]interface{}{
					"integration_id": config["integration_id"],
					"channel":        config["channel"],
					"channel_id":     config["channel_id"],
				}
			}
			if err := d.Set("slack_config", slackConfigs); err != nil {
				return fmt.Errorf("failed to set slack_config: %w", err)
			}
		}
	}

	return nil
}

func getNotificationChannelTypeAndConfig(d *schema.ResourceData) (client.NotificationChannelType, []map[string]interface{}, error) {
	if emailConfigs, ok := d.GetOk("email_config"); ok {
		configList := emailConfigs.([]interface{})
		if len(configList) > 0 {
			result := make([]map[string]interface{}, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]interface{})
				result[i] = map[string]interface{}{
					"address": configMap["address"],
				}
			}
			return client.NotificationChannelTypeEmail, result, nil
		}
	}

	if webhookConfigs, ok := d.GetOk("webhook_config"); ok {
		configList := webhookConfigs.([]interface{})
		if len(configList) > 0 {
			result := make([]map[string]interface{}, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]interface{})
				result[i] = map[string]interface{}{
					"url": configMap["url"],
				}
				if method, ok := configMap["method"]; ok && method != "" {
					result[i]["method"] = method
				}
			}
			return client.NotificationChannelTypeWebhook, result, nil
		}
	}

	if slackConfigs, ok := d.GetOk("slack_config"); ok {
		configList := slackConfigs.([]interface{})
		if len(configList) > 0 {
			result := make([]map[string]interface{}, len(configList))
			for i, configInterface := range configList {
				configMap := configInterface.(map[string]interface{})
				result[i] = map[string]interface{}{
					"integration_id": configMap["integration_id"],
					"channel":        configMap["channel"],
				}
			}
			return client.NotificationChannelTypeSlack, result, nil
		}
	}

	return "", nil, fmt.Errorf("exactly one of email_config, webhook_config, or slack_config must be specified")
}
