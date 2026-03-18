package twilio_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about a Twilio access privilege in the Hush Security platform.",
		ReadContext: resourceRead,
		Schema:      DataSourceSchema(),
	}
}
