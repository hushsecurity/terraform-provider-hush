package mariadb_access_credential

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about a MariaDB access credential in the Hush Security platform.",
		ReadContext: resourceRead,
		Schema:      DataSourceSchema(),
	}
}
