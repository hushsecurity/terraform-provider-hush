package deployment

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

// testAccProvider is a test provider instance for acceptance tests
var testAccProvider *schema.Provider

func init() {
	testAccProvider = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HUSH_API_KEY_ID", nil),
			},
			"api_key_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("HUSH_API_KEY_SECRET", nil),
			},
			"realm": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HUSH_REALM", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hush_deployment": Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hush_deployment": DataSource(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	client, err := client.NewClient(
		ctx,
		d.Get("api_key_id").(string),
		d.Get("api_key_secret").(string),
		d.Get("realm").(string),
	)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return client, nil
}

var providerFactories = map[string]func() (*schema.Provider, error){
	"hush": func() (*schema.Provider, error) {
		return testAccProvider, nil
	},
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("HUSH_API_KEY_ID") == "" {
		t.Fatalf("HUSH_API_KEY_ID env var must be set")
	}
	if os.Getenv("HUSH_API_KEY_SECRET") == "" {
		t.Fatalf("HUSH_API_KEY_SECRET env var must be set")
	}
	if os.Getenv("HUSH_REALM") == "" {
		t.Fatalf("HUSH_REALM env var must be set")
	}
}

func TestAccResourceDeployment_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			// Create step
			{
				Config: testAccDeploymentConfig_basic("test-deployment", "test deployment description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists("hush_deployment.test"),
					resource.TestMatchResourceAttr(
						"hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "name", "test-deployment",
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "description", "test deployment description",
					),
				),
			},
			// Update step
			{
				Config: testAccDeploymentConfig_basic("test-deployment-updated", "updated deployment description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists("hush_deployment.test"),
					resource.TestMatchResourceAttr(
						"hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "name", "test-deployment-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "description", "updated deployment description",
					),
				),
			},
			// Import step
			{
				ResourceName:      "hush_deployment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSourceDeployment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic("test-deployment", "test deployment description") + testAccDeploymentDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_deployment.test", "name", "test-deployment",
					),
					resource.TestCheckResourceAttr(
						"data.hush_deployment.test", "description", "test deployment description",
					),
				),
			},
		},
	})
}

// testAccCheckDeploymentDestroy verifies all deployment resources have been destroyed
func testAccCheckDeploymentDestroy(s *terraform.State) error {
	c, err := client.NewClient(
		context.Background(),
		os.Getenv("HUSH_API_KEY_ID"),
		os.Getenv("HUSH_API_KEY_SECRET"),
		os.Getenv("HUSH_REALM"),
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hush_deployment" {
			continue
		}

		_, err := client.GetDeployment(context.Background(), c, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("deployment %s still exists", rs.Primary.ID)
		}

		apiError, ok := err.(*client.APIError)
		if ok && apiError.IsNotFound() {
			continue // Resource properly destroyed
		}
		return fmt.Errorf("failed to verify deployment %s was destroyed: %s", rs.Primary.ID, err)
	}
	return nil
}

// testAccCheckDeploymentExists verifies a deployment resource exists in the API
func testAccCheckDeploymentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for resource: %s", resourceName)
		}

		c, err := client.NewClient(
			context.Background(),
			os.Getenv("HUSH_API_KEY_ID"),
			os.Getenv("HUSH_API_KEY_SECRET"),
			os.Getenv("HUSH_REALM"),
		)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		_, err = client.GetDeployment(context.Background(), c, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("deployment %s not found: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

// testAccDeploymentConfig_basic returns a basic deployment configuration
func testAccDeploymentConfig_basic(name, description string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

// testAccDeploymentDataSourceConfig returns a data source configuration
const testAccDeploymentDataSourceConfig = `
data "hush_deployment" "test" {
  id = hush_deployment.test.id
}
`
