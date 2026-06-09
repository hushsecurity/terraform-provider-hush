package artifactory_integration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const dataSourceDescription = "Use this data source to read information about an existing Hush Security Artifactory integration."

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,

		ReadContext: artifactoryIntegrationRead,
		Schema:      ArtifactoryIntegrationDataSourceSchema(),
	}
}
