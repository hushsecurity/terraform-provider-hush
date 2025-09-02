package notification_configuration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "Notification configuration data source for reading Hush Security notification configuration information"

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: notificationConfigurationRead,
		Schema:      NotificationConfigurationDataSourceSchema(),
	}
}
