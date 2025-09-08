package notification_configuration

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
			"hush_notification_configuration": Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hush_notification_configuration": DataSource(),
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

func TestAccResourceNotificationConfiguration_basic(t *testing.T) {
	t.Skip("Notification configuration API returns 'Method Not Allowed' - implementation pending")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			// Create step
			{
				Config: testAccNotificationConfigurationConfig_basic("test-config", "test configuration description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationConfigurationExists("hush_notification_configuration.test"),
					resource.TestMatchResourceAttr(
						"hush_notification_configuration.test", "id", regexp.MustCompile("^ncf-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.test", "name", "test-config",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.test", "description", "test configuration description",
					),
				),
			},
			// Update step
			{
				Config: testAccNotificationConfigurationConfig_basic("test-config-updated", "updated test configuration description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationConfigurationExists("hush_notification_configuration.test"),
					resource.TestMatchResourceAttr(
						"hush_notification_configuration.test", "id", regexp.MustCompile("^ncf-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.test", "name", "test-config-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.test", "description", "updated test configuration description",
					),
				),
			},
			// Import step
			{
				ResourceName:      "hush_notification_configuration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDataSourceNotificationConfiguration(t *testing.T) {
	t.Skip("Notification configuration API returns 'Method Not Allowed' - implementation pending")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckNotificationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_basic("test-config", "test configuration description") + testAccNotificationConfigurationDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_notification_configuration.config", "id", regexp.MustCompile("^ncf-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "name", "test-config",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "description", "test configuration description",
					),
				),
			},
		},
	})
}

// testAccCheckNotificationConfigurationDestroy verifies all notification configuration resources have been destroyed
func testAccCheckNotificationConfigurationDestroy(s *terraform.State) error {
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
		if rs.Type != "hush_notification_configuration" {
			continue
		}

		_, err := client.GetNotificationConfiguration(context.Background(), c, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("notification configuration %s still exists", rs.Primary.ID)
		}

		apiError, ok := err.(*client.APIError)
		if ok && apiError.IsNotFound() {
			continue // Resource properly destroyed
		}
		return fmt.Errorf("failed to verify notification configuration %s was destroyed: %s", rs.Primary.ID, err)
	}
	return nil
}

// testAccCheckNotificationConfigurationExists verifies a notification configuration resource exists in the API
func testAccCheckNotificationConfigurationExists(resourceName string) resource.TestCheckFunc {
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

		_, err = client.GetNotificationConfiguration(context.Background(), c, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("notification configuration %s not found: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

// testAccNotificationConfigurationConfig_basic returns a basic notification configuration
func testAccNotificationConfigurationConfig_basic(name, description string) string {
	return fmt.Sprintf(`
resource "hush_notification_configuration" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

// testAccNotificationConfigurationDataSourceConfig returns a data source configuration
const testAccNotificationConfigurationDataSourceConfig = `
data "hush_notification_configuration" "config" {
  id = hush_notification_configuration.test.id
}
`
