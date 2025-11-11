package access_policy

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "Access policy data source for reading Hush Security access policy information"

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: accessPolicyRead,
		Schema:      AccessPolicyDataSourceSchema(),
	}
}
