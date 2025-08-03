package deployment

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "Deployment data source for reading Hush Security deployment information"

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: deploymentRead,
		Schema:      DeploymentDataSourceSchema(),
	}
}
