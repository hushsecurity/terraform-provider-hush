package notification_channel

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Notification channel resource for managing Hush Security notification channels"

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: notificationChannelCreate,
		ReadContext:   notificationChannelRead,
		UpdateContext: notificationChannelUpdate,
		DeleteContext: notificationChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: NotificationChannelResourceSchema(),
	}
}

func notificationChannelCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	_, config, err := getNotificationChannelTypeAndConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	input := &client.CreateNotificationChannelInput{
		Name:    d.Get("name").(string),
		Enabled: d.Get("enabled").(bool),
		Config:  config,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = &desc
	}

	resp, err := client.CreateNotificationChannel(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return notificationChannelRead(ctx, d, m)
}

func notificationChannelUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateNotificationChannelInput{}
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
	if d.HasChanges("email_config", "webhook_config", "slack_config") {
		_, config, err := getNotificationChannelTypeAndConfig(d)
		if err != nil {
			return diag.FromErr(err)
		}
		input.Config = &config
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	_, err := client.UpdateNotificationChannel(ctx, c, d.Id(), input)
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return notificationChannelRead(ctx, d, m)
}

func notificationChannelDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteNotificationChannel(ctx, c, d.Id())
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
