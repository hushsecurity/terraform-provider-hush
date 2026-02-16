package gcp_integration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "GCP integration data source for reading Hush Security GCP integration information"

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: gcpIntegrationRead,
		Schema:      GCPIntegrationDataSourceSchema(),
	}
}
