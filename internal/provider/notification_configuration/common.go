package notification_configuration

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
	idDesc              = "The unique identifier of the notification configuration"
	nameDesc            = "The name of the notification configuration"
	descriptionDesc     = "The description of the notification configuration"
	enabledDesc         = "Whether the notification configuration is enabled"
	channelIDsDesc      = "The list of notification channel IDs associated with this configuration"
	aggregationDesc     = "The aggregation duration for notifications (short, week, month)"
	triggerDesc         = "The trigger type for notifications (new_nhi_at_risk, nhi_digest)"
	createdAtDesc       = "The creation timestamp of the notification configuration"
	modifiedAtDesc      = "The last modification timestamp of the notification configuration"
	lastTriggeredAtDesc = "The last trigger timestamp of the notification configuration"
)

func NotificationConfigurationResourceSchema() map[string]*schema.Schema {
	s := NotificationConfigurationDataSourceSchema()

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
	s["channel_ids"] = &schema.Schema{
		Description: channelIDsDesc,
		Type:        schema.TypeList,
		Required:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	s["aggregation"] = &schema.Schema{
		Description: aggregationDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringInSlice([]string{
			string(client.AggregationDurationShort),
			string(client.AggregationDurationWeek),
			string(client.AggregationDurationMonth),
		}, false),
	}
	s["trigger"] = &schema.Schema{
		Description: triggerDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringInSlice([]string{
			string(client.TriggerNewNHIAtRisk),
			string(client.TriggerNHIDigest),
		}, false),
	}

	return s
}

func NotificationConfigurationDataSourceSchema() map[string]*schema.Schema {
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
		"channel_ids": {
			Description: channelIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"aggregation": {
			Description: aggregationDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"trigger": {
			Description: triggerDesc,
			Type:        schema.TypeString,
			Computed:    true,
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
		"last_triggered_at": {
			Description: lastTriggeredAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

// Helper Functions

func notificationConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	var config *client.NotificationConfiguration
	var err error

	if id := d.Id(); id != "" {
		config, err = client.GetNotificationConfiguration(ctx, c, id)
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
		configID := id.(string)
		config, err = client.GetNotificationConfiguration(ctx, c, configID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no notification configuration found with ID: %s", configID)
			} else {
				return diag.FromErr(err)
			}
		}
	} else if name, exists := d.GetOk("name"); exists {
		// Lookup by name
		configName := name.(string)
		configs, err := client.GetNotificationConfigurationsByName(ctx, c, configName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup notification configuration by name '%s': %w", configName, err))
		}

		switch len(configs) {
		case 0:
			return diag.Errorf("no notification configuration found with name: %s", configName)
		case 1:
			config = &configs[0]
		default:
			return diag.Errorf("multiple notification configurations found with name '%s'. Configuration names must be unique. Consider using the configuration ID instead for exact matching", configName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(config.ID)
	}

	if diags := setNotificationConfigurationFields(d, config); diags.HasError() {
		return diags
	}

	return nil
}

func setNotificationConfigurationFields(d *schema.ResourceData, config *client.NotificationConfiguration) diag.Diagnostics {
	fields := map[string]interface{}{
		"name":              config.Name,
		"description":       config.Description,
		"enabled":           config.Enabled,
		"channel_ids":       config.ChannelIDs,
		"aggregation":       string(config.Aggregation),
		"trigger":           string(config.Trigger),
		"created_at":        config.CreatedAt,
		"modified_at":       config.ModifiedAt,
		"last_triggered_at": config.LastTriggeredAt,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}

func buildNotificationConfigurationChannelIDs(d *schema.ResourceData) []string {
	channelIDsRaw := d.Get("channel_ids").([]interface{})
	channelIDs := make([]string, len(channelIDsRaw))
	for i, v := range channelIDsRaw {
		channelIDs[i] = v.(string)
	}
	return channelIDs
}
