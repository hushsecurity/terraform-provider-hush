package notification_configuration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Notification configuration resource for managing Hush Security notification configurations. " +
	"With config_id it manages a predefined configuration; without it, it creates one, such as an activity stream (trigger agw_activity). " +
	"An imported configuration is disabled on destroy, not deleted; a predefined one is adopted with config_id and needs no import."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: notificationConfigurationCreate,
		ReadContext:   resourceRead,
		UpdateContext: notificationConfigurationUpdate,
		DeleteContext: notificationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: ValidateMode,
		Schema:        NotificationConfigurationResourceSchema(),
	}
}

func predefined(d *schema.ResourceData) bool {
	return d.Get("config_id").(string) != ""
}

func notificationConfigurationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if predefined(d) {
		// patched before reading: a read first would replace the configured
		// channel_ids with the stored ones, and they would never be sent
		d.SetId(d.Get("config_id").(string))
		enabled := d.Get("enabled").(bool)
		channelIDs := buildNotificationConfigurationChannelIDs(d)
		input := &client.UpdateNotificationConfigurationInput{Enabled: &enabled, ChannelIDs: &channelIDs}
		if v, ok := d.GetOk("aggregation"); ok {
			aggregation := client.AggregationDuration(v.(string))
			input.Aggregation = &aggregation
		}
		if _, err := client.UpdateNotificationConfiguration(ctx, c, d.Id(), input); err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
		return resourceRead(ctx, d, m)
	}

	filters, err := buildDispatchFilters(d)
	if err != nil {
		return diag.FromErr(err)
	}
	input := &client.CreateNotificationConfigurationInput{
		Name:            d.Get("name").(string),
		Enabled:         d.Get("enabled").(bool),
		ChannelIDs:      buildNotificationConfigurationChannelIDs(d),
		Trigger:         client.Trigger(d.Get("trigger").(string)),
		DispatchFilters: filters,
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		input.Description = &description
	}
	if v, ok := d.GetOk("aggregation"); ok {
		aggregation := client.AggregationDuration(v.(string))
		input.Aggregation = &aggregation
	}
	created, err := client.CreateNotificationConfiguration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(created.ID)
	if err := d.Set("created", true); err != nil {
		return diag.FromErr(err)
	}
	return resourceRead(ctx, d, m)
}

func notificationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateNotificationConfigurationInput{}
	hasChanges := false

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		input.Enabled = &enabled
		hasChanges = true
	}
	if d.HasChange("channel_ids") {
		channelIDs := buildNotificationConfigurationChannelIDs(d)
		input.ChannelIDs = &channelIDs
		hasChanges = true
	}
	if d.HasChange("aggregation") {
		aggregation := client.AggregationDuration(d.Get("aggregation").(string))
		input.Aggregation = &aggregation
		hasChanges = true
	}
	if !predefined(d) {
		if d.HasChange("name") {
			name := d.Get("name").(string)
			input.Name = &name
			hasChanges = true
		}
		if d.HasChange("description") {
			description := d.Get("description").(string)
			input.Description = &description
			hasChanges = true
		}
		if d.HasChange("dispatch_filters") {
			filters, err := buildDispatchFilters(d)
			if err != nil {
				return diag.FromErr(err)
			}
			input.DispatchFilters = &filters
			hasChanges = true
		}
	}

	if !hasChanges {
		return nil
	}

	_, err := client.UpdateNotificationConfiguration(ctx, c, d.Id(), input)
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, m)
}

func notificationConfigurationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var err error
	if !d.Get("created").(bool) {
		// predefined or imported: reset, never deleted
		_, err = client.UpdateNotificationConfiguration(ctx, c, d.Id(), &client.UpdateNotificationConfigurationInput{
			Enabled:    &[]bool{false}[0],
			ChannelIDs: &[]string{},
		})
	} else {
		err = client.DeleteNotificationConfiguration(ctx, c, d.Id())
	}
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if !ok || errResponse.StatusCode != http.StatusNotFound {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}
