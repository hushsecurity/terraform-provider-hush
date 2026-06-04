package bitbucket_integration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "Use this data source to read information about an existing Hush Security Bitbucket integration."

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: bitbucketIntegrationRead,
		Schema:      BitbucketIntegrationDataSourceSchema(),
	}
}
