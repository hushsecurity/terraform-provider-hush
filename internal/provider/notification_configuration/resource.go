package notification_configuration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Notification configuration resource for managing Hush Security notification configurations"

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: notificationConfigurationCreate,
		ReadContext:   notificationConfigurationRead,
		UpdateContext: notificationConfigurationUpdate,
		DeleteContext: notificationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: NotificationConfigurationResourceSchema(),
	}
}

func notificationConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateNotificationConfigurationInput{
		Name:        d.Get("name").(string),
		Enabled:     d.Get("enabled").(bool),
		ChannelIDs:  buildNotificationConfigurationChannelIDs(d),
		Aggregation: client.AggregationDuration(d.Get("aggregation").(string)),
		Trigger:     client.Trigger(d.Get("trigger").(string)),
	}

	// Only include description if it's not empty
	if desc := d.Get("description").(string); desc != "" {
		input.Description = &desc
	}

	resp, err := client.CreateNotificationConfiguration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return notificationConfigurationRead(ctx, d, m)
}

func notificationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateNotificationConfigurationInput{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
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
	if d.HasChange("trigger") {
		trigger := client.Trigger(d.Get("trigger").(string))
		input.Trigger = &trigger
		hasChanges = true
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

	return notificationConfigurationRead(ctx, d, m)
}

func notificationConfigurationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteNotificationConfiguration(ctx, c, d.Id())
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
