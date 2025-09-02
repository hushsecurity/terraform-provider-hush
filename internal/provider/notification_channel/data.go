package notification_channel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "Notification channel data source for reading Hush Security notification channel information"

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: notificationChannelRead,
		Schema:      NotificationChannelDataSourceSchema(),
	}
}
