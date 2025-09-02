package notification_configuration

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc              = "The unique identifier of the notification configuration"
	nameDesc            = "The name of the notification configuration"
	descriptionDesc     = "The description of the notification configuration"
	enabledDesc         = "Whether the notification configuration is enabled"
	channelIDsDesc      = "The list of notification channel IDs to send notifications to"
	aggregationDesc     = "The aggregation duration for notifications"
	triggerDesc         = "The trigger type for notifications"
	createdAtDesc       = "The creation timestamp of the notification configuration"
	modifiedAtDesc      = "The last modification timestamp of the notification configuration"
	lastTriggeredAtDesc = "The last trigger timestamp of the notification configuration"
)

func NotificationConfigurationResourceSchema() map[string]*schema.Schema {
	s := NotificationConfigurationDataSourceSchema()

	s["config_id"] = &schema.Schema{
		Description: "The ID of the predefined notification configuration to manage",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}

	s["name"] = &schema.Schema{
		Description: nameDesc + " (read-only)",
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["description"] = &schema.Schema{
		Description: descriptionDesc + " (read-only)",
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["aggregation"] = &schema.Schema{
		Description: aggregationDesc + " (read-only)",
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["trigger"] = &schema.Schema{
		Description: triggerDesc + " (read-only)",
		Type:        schema.TypeString,
		Computed:    true,
	}

	s["enabled"] = &schema.Schema{
		Description: enabledDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	}
	s["channel_ids"] = &schema.Schema{
		Description: channelIDsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
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
			ConflictsWith: []string{"name", "trigger"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"id", "trigger"},
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
			Description:   triggerDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"id", "name"},
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
	} else if trigger, exists := d.GetOk("trigger"); exists {
		// Lookup by trigger
		triggerType := trigger.(string)
		configs, err := client.GetNotificationConfigurationsByTrigger(ctx, c, triggerType)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup notification configuration by trigger '%s': %w", triggerType, err))
		}

		switch len(configs) {
		case 0:
			return diag.Errorf("no notification configuration found with trigger: %s", triggerType)
		case 1:
			config = &configs[0]
		default:
			return diag.Errorf("multiple notification configurations found with trigger '%s'. Consider using the configuration ID instead for exact matching", triggerType)
		}
	} else {
		return diag.Errorf("either 'id', 'name', or 'trigger' must be specified")
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
