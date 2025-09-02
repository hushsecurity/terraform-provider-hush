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
	configID := d.Get("config_id").(string)
	if configID == "" {
		return diag.Errorf("config_id is required for notification configuration resources. Use a data source to look up predefined configuration IDs.")
	}

	d.SetId(configID)

	readDiags := notificationConfigurationRead(ctx, d, m)
	if readDiags.HasError() {
		return readDiags
	}

	return notificationConfigurationUpdate(ctx, d, m)
}

func notificationConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	input := &client.UpdateNotificationConfigurationInput{
		Enabled:    &[]bool{false}[0],
		ChannelIDs: &[]string{},
	}

	_, err := client.UpdateNotificationConfiguration(ctx, c, d.Id(), input)
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}
