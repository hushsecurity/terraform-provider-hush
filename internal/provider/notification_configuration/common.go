package notification_configuration

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
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
	channelIDsDesc      = "The list of notification channel IDs to send notifications to"
	aggregationDesc     = "How notifications are grouped: short, day, week, month, or stream for an activity stream"
	triggerDesc         = "What the notifications are about"
	dispatchFiltersDesc = "Filters narrowing what is sent. Which names a trigger takes is the API's to check."
	statusDesc          = "ok, warning, error or disabled: whether what it selected can be delivered to"
	statusMessageDesc   = "Why the status is not ok, when it is not"
	lastTriggeredAtDesc = "The last trigger timestamp of the notification configuration"
)

// creatableTriggers are the ones the API creates; nhi_digest is predefined only.
var creatableTriggers = []string{
	string(client.TriggerNewNHISecurityIssue),
	string(client.TriggerNewAIAgentSecurityIssue),
	string(client.TriggerCertificateExpiring),
	string(client.TriggerAGWActivity),
}

func NotificationConfigurationResourceSchema() map[string]*schema.Schema {
	s := NotificationConfigurationDataSourceSchema()

	s["config_id"] = &schema.Schema{
		Description: "The ID of a predefined notification configuration to manage. Without it, the resource creates one from name and trigger.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	}
	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	// Optional and computed: set them to create a configuration, or leave them
	// to read a predefined one's.
	s["name"] = &schema.Schema{
		Description: nameDesc + ". Required without config_id.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	}
	s["description"] = &schema.Schema{
		Description: descriptionDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	}
	s["trigger"] = &schema.Schema{
		Description:  triggerDesc + ". Required without config_id; changing it creates a new configuration.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(creatableTriggers, false),
	}
	s["aggregation"] = &schema.Schema{
		Description: aggregationDesc + ". Set only where the trigger takes one: certificate_expiring (day or week) and nhi_digest (day, week or month); the others are fixed.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ValidateFunc: validation.StringInSlice([]string{
			string(client.AggregationDurationDay),
			string(client.AggregationDurationWeek),
			string(client.AggregationDurationMonth),
		}, false),
	}
	s["dispatch_filters"] = &schema.Schema{
		Description: dispatchFiltersDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "severity, env_type, notification_type, expires_in_days or deployment",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: validation.StringInSlice([]string{
						"severity", "env_type", "notification_type", "expires_in_days", "deployment",
					}, false),
				},
				"op": {
					Description:  "eq or neq; gte or lte for expires_in_days",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"eq", "neq", "gte", "lte"}, false),
				},
				"values": {
					Description: "The values to match. expires_in_days takes one number of days.",
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
	s["enabled"] = &schema.Schema{
		Description: enabledDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	}
	s["created"] = &schema.Schema{
		Description: "Whether this resource created the configuration. Only then does destroy delete it; otherwise, as for a predefined or imported one, destroy disables it and clears its channels.",
		Type:        schema.TypeBool,
		Computed:    true,
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
		"dispatch_filters": {
			Description: dispatchFiltersDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name":   {Description: "The filter's name", Type: schema.TypeString, Computed: true},
					"op":     {Description: "eq, neq, gte or lte", Type: schema.TypeString, Computed: true},
					"values": {Description: "The values it matches", Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				},
			},
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status_message": {
			Description: statusMessageDesc,
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

// ValidateMode checks at plan what the schema can't: a created configuration
// needs a name and a trigger, a predefined one takes neither, and aggregation
// goes only where the trigger takes one.
func ValidateMode(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	raw := diff.GetRawConfig()
	if raw.IsNull() {
		return nil
	}
	// unknown until apply counts as set; an absent block or "" does not
	set := func(attr string) bool {
		v := raw.GetAttr(attr)
		switch {
		case v.IsNull():
			return false
		case !v.IsKnown():
			return true
		case v.Type().IsListType():
			return v.LengthInt() > 0
		case v.Type() == cty.String:
			return v.AsString() != ""
		}
		return true
	}
	if set("config_id") {
		for _, attr := range []string{"name", "trigger", "description", "dispatch_filters"} {
			if set(attr) {
				return fmt.Errorf("%s is the predefined configuration's own: it can't be set with config_id", attr)
			}
		}
		return nil
	}
	if !set("name") || !set("trigger") {
		return fmt.Errorf("name and trigger are required without config_id")
	}
	if !raw.GetAttr("trigger").IsKnown() {
		return nil
	}
	trigger := diff.Get("trigger").(string)
	switch {
	case trigger == string(client.TriggerCertificateExpiring) && !set("aggregation"):
		return fmt.Errorf("aggregation (day or week) is required for trigger %s", trigger)
	case trigger != string(client.TriggerCertificateExpiring) && set("aggregation"):
		return fmt.Errorf("aggregation is fixed for trigger %s: leave it unset", trigger)
	case trigger == string(client.TriggerCertificateExpiring) && diff.Get("aggregation").(string) == string(client.AggregationDurationMonth):
		return fmt.Errorf("aggregation for trigger %s is day or week", trigger)
	}
	return nil
}

// Helper Functions

func notificationConfigurationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return readConfiguration(ctx, d, m, true)
}

// resourceRead leaves a predefined configuration's filters alone: they are its
// own, and reading them in would plan their removal.
func resourceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return readConfiguration(ctx, d, m, !predefined(d))
}

func readConfiguration(ctx context.Context, d *schema.ResourceData, m any, withFilters bool) diag.Diagnostics {
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

	if diags := setNotificationConfigurationFields(d, config, withFilters); diags.HasError() {
		return diags
	}

	return nil
}

func setNotificationConfigurationFields(d *schema.ResourceData, config *client.NotificationConfiguration, withFilters bool) diag.Diagnostics {
	fields := map[string]any{
		"name":              config.Name,
		"description":       config.Description,
		"enabled":           config.Enabled,
		"channel_ids":       config.ChannelIDs,
		"aggregation":       string(config.Aggregation),
		"trigger":           string(config.Trigger),
		"status":            config.Status,
		"status_message":    config.StatusMessage,
		"last_triggered_at": config.LastTriggeredAt,
	}

	if withFilters {
		fields["dispatch_filters"] = flattenDispatchFilters(config.DispatchFilters)
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}

func flattenDispatchFilters(filters []client.DispatchFilter) []map[string]any {
	out := make([]map[string]any, len(filters))
	for i, f := range filters {
		values := make([]string, len(f.Values))
		for j, v := range f.Values {
			// a number reads back as 7, not 7.000000
			if n, ok := v.(float64); ok {
				values[j] = strconv.FormatFloat(n, 'f', -1, 64)
			} else {
				values[j] = fmt.Sprint(v)
			}
		}
		out[i] = map[string]any{"name": f.Name, "op": f.Op, "values": values}
	}
	return out
}

func buildDispatchFilters(d *schema.ResourceData) ([]client.DispatchFilter, error) {
	raw, _ := d.Get("dispatch_filters").([]any)
	filters := make([]client.DispatchFilter, 0, len(raw))
	for i, item := range raw {
		f := item.(map[string]any)
		name := f["name"].(string)
		var values []any
		for _, v := range f["values"].([]any) {
			s := v.(string)
			if name != "expires_in_days" {
				values = append(values, s)
				continue
			}
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("dispatch_filters[%d]: expires_in_days takes a number, got %q", i, s)
			}
			values = append(values, n)
		}
		filters = append(filters, client.DispatchFilter{Name: name, Op: f["op"].(string), Values: values})
	}
	return filters, nil
}

func buildNotificationConfigurationChannelIDs(d *schema.ResourceData) []string {
	channelIDsRaw := d.Get("channel_ids").([]any)
	channelIDs := make([]string, len(channelIDsRaw))
	for i, v := range channelIDsRaw {
		channelIDs[i] = v.(string)
	}
	return channelIDs
}
