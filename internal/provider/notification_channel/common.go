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
	idDesc                = "The unique identifier of the notification channel"
	nameDesc              = "The name of the notification channel"
	descriptionDesc       = "The description of the notification channel"
	enabledDesc           = "Whether the notification channel is enabled"
	typeDesc              = "The type of the notification channel (email, webhook, slack)"
	configDesc            = "The configuration for the notification channel"
	createdAtDesc         = "The creation timestamp of the notification channel"
	modifiedAtDesc        = "The last modification timestamp of the notification channel"
	configAddressDesc     = "The email address for email notifications"
	configVerifiedDesc    = "Whether the email address or webhook URL is verified"
	configURLDesc         = "The webhook URL"
	configMethodDesc      = "The HTTP method for webhook requests (POST, GET)"
	configIntegrationDesc = "The Slack integration ID"
	configChannelDesc     = "The Slack channel name"
	configChannelIDDesc   = "The Slack channel ID"
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
		Required:    true,
	}
	s["type"] = &schema.Schema{
		Description: typeDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringInSlice([]string{
			string(client.NotificationChannelTypeEmail),
			string(client.NotificationChannelTypeWebhook),
			string(client.NotificationChannelTypeSlack),
		}, false),
	}
	s["config"] = &schema.Schema{
		Description: configDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address": {
					Description: configAddressDesc,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"verified": {
					Description: configVerifiedDesc,
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"url": {
					Description: configURLDesc,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"method": {
					Description: configMethodDesc,
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "POST",
					ValidateFunc: validation.StringInSlice([]string{
						string(client.WebhookMethodPOST),
						string(client.WebhookMethodGET),
					}, false),
				},
				"integration_id": {
					Description: configIntegrationDesc,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"channel": {
					Description: configChannelDesc,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"channel_id": {
					Description: configChannelIDDesc,
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
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"config": {
			Description: configDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"address": {
						Description: configAddressDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"verified": {
						Description: configVerifiedDesc,
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"url": {
						Description: configURLDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"method": {
						Description: configMethodDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"integration_id": {
						Description: configIntegrationDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"channel": {
						Description: configChannelDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"channel_id": {
						Description: configChannelIDDesc,
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
		// Lookup by ID provided in configuration
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
		// Lookup by name
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

	// Set config field
	if err := d.Set("config", channel.Config); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set config: %w", err))
	}

	return nil
}

func buildNotificationChannelConfig(d *schema.ResourceData) []map[string]interface{} {
	configList := d.Get("config").([]interface{})
	channelType := d.Get("type").(string)
	result := make([]map[string]interface{}, len(configList))

	for i, configItem := range configList {
		configMap := configItem.(map[string]interface{})
		result[i] = make(map[string]interface{})

		// Only include fields that the API accepts based on the Herald API guide
		switch channelType {
		case "email":
			// EmailConfigIn: Only "address" field allowed
			if address := configMap["address"]; address != nil && address != "" {
				result[i]["address"] = address
			}
		case "webhook":
			// WebhookConfigIn: Only "url" and optional "method" fields allowed
			if url := configMap["url"]; url != nil && url != "" {
				result[i]["url"] = url
			}
			// Method is optional, defaults to POST on backend
			if method := configMap["method"]; method != nil && method != "" {
				result[i]["method"] = method
			}
		case "slack":
			// SlackConfigIn: Only "integration_id" and "channel" fields allowed
			if integrationID := configMap["integration_id"]; integrationID != nil && integrationID != "" {
				result[i]["integration_id"] = integrationID
			}
			if channel := configMap["channel"]; channel != nil && channel != "" {
				result[i]["channel"] = channel
			}
		}
	}

	return result
}
