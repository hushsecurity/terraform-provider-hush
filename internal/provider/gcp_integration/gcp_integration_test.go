package gcp_integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
			"hush_gcp_integration": Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hush_gcp_integration": DataSource(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	realm := strings.ToLower(d.Get("realm").(string))

	var baseURL string
	if devURL := os.Getenv("HUSH_DEV_BASE_URL"); devURL != "" {
		baseURL = devURL
	} else {
		baseURL = fmt.Sprintf("https://api.%s.hush-security.com", realm)
	}

	c, err := client.NewClient(
		ctx,
		d.Get("api_key_id").(string),
		d.Get("api_key_secret").(string),
		baseURL,
	)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return c, nil
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

// --- Resource Tests ---

func TestAccResourceGCPIntegration_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckGCPIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create step
			{
				Config: testAccGCPIntegrationConfig_basic("test-gcp-integration", "test gcp integration description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCPIntegrationExists("hush_gcp_integration.test"),
					resource.TestCheckResourceAttrSet(
						"hush_gcp_integration.test", "id",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "name", "test-gcp-integration",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "description", "test gcp integration description",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "status", "pending",
					),
				),
			},
			// Update step
			{
				Config: testAccGCPIntegrationConfig_basic("test-gcp-integration-updated", "updated gcp integration description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCPIntegrationExists("hush_gcp_integration.test"),
					resource.TestCheckResourceAttrSet(
						"hush_gcp_integration.test", "id",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "name", "test-gcp-integration-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "description", "updated gcp integration description",
					),
				),
			},
			// Import step
			{
				ResourceName:            "hush_gcp_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"onboarding_script", "feature", "project"},
			},
		},
	})
}

func TestAccResourceGCPIntegration_withProjects(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckGCPIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create with projects
			{
				Config: testAccGCPIntegrationConfig_withProjects("test-gcp-with-projects", "integration with projects", "my-gcp-project-01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCPIntegrationExists("hush_gcp_integration.test"),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "name", "test-gcp-with-projects",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "project.0.project_id", "my-gcp-project-01",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "project.0.enabled", "true",
					),
				),
			},
			// Import step
			{
				ResourceName:            "hush_gcp_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"onboarding_script", "feature", "project"},
			},
		},
	})
}

func TestAccResourceGCPIntegration_withFeatures(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckGCPIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create with features
			{
				Config: testAccGCPIntegrationConfig_withFeatures("test-gcp-with-features", "integration with features"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCPIntegrationExists("hush_gcp_integration.test"),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "name", "test-gcp-with-features",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.0.name", "iam",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.0.enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.1.name", "secret_manager",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.1.enabled", "true",
					),
				),
			},
			// Import step
			{
				ResourceName:            "hush_gcp_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"onboarding_script", "feature", "project"},
			},
		},
	})
}

func TestAccResourceGCPIntegration_full(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckGCPIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create with projects and features
			{
				Config: testAccGCPIntegrationConfig_full("test-gcp-full", "full integration", "my-gcp-project-01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCPIntegrationExists("hush_gcp_integration.test"),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "name", "test-gcp-full",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "description", "full integration",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "project.0.project_id", "my-gcp-project-01",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.0.name", "iam",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.0.enabled", "true",
					),
				),
			},
			// Update - change description and add another feature
			{
				Config: testAccGCPIntegrationConfig_fullUpdated("test-gcp-full", "updated full integration", "my-gcp-project-01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCPIntegrationExists("hush_gcp_integration.test"),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "description", "updated full integration",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.0.name", "iam",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_integration.test", "feature.1.name", "secret_manager",
					),
				),
			},
			// Import step
			{
				ResourceName:            "hush_gcp_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"onboarding_script", "feature", "project"},
			},
		},
	})
}

// --- Data Source Tests ---

func TestAccDataSourceGCPIntegration_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckGCPIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPIntegrationConfig_basic("test-gcp-ds-id", "data source test by id") + testAccGCPIntegrationDataSourceConfig_byID,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.hush_gcp_integration.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_integration.test", "name", "test-gcp-ds-id",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_integration.test", "description", "data source test by id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGCPIntegration_byName(t *testing.T) {
	// Use a unique name to avoid collisions with stale test data
	uniqueName := fmt.Sprintf("test-gcp-ds-name-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckGCPIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGCPIntegrationConfig_basic(uniqueName, "data source test by name") + testAccGCPIntegrationDataSourceConfig_byName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.hush_gcp_integration.test", "id",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_integration.test", "name", uniqueName,
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_integration.test", "description", "data source test by name",
					),
				),
			},
		},
	})
}

// --- Destroy and Exists check functions ---

// testAccCheckGCPIntegrationDestroy verifies all GCP integration resources have been destroyed
func testAccCheckGCPIntegrationDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hush_gcp_integration" {
			continue
		}

		_, err := client.GetGCPIntegration(context.Background(), c, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("GCP integration %s still exists", rs.Primary.ID)
		}

		apiError, ok := err.(*client.APIError)
		if ok && apiError.IsNotFound() {
			continue // Resource properly destroyed
		}
		return fmt.Errorf("failed to verify GCP integration %s was destroyed: %s", rs.Primary.ID, err)
	}
	return nil
}

// testAccCheckGCPIntegrationExists verifies a GCP integration resource exists in the API
func testAccCheckGCPIntegrationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for resource: %s", resourceName)
		}

		c := testAccProvider.Meta().(*client.Client)

		_, err := client.GetGCPIntegration(context.Background(), c, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("GCP integration %s not found: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

// --- Test configuration helpers ---

// testAccGCPIntegrationConfig_basic returns a basic GCP integration configuration
func testAccGCPIntegrationConfig_basic(name, description string) string {
	return fmt.Sprintf(`
resource "hush_gcp_integration" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

// testAccGCPIntegrationConfig_withProjects returns a GCP integration with a project
func testAccGCPIntegrationConfig_withProjects(name, description, projectID string) string {
	return fmt.Sprintf(`
resource "hush_gcp_integration" "test" {
  name        = %[1]q
  description = %[2]q

  project {
    project_id = %[3]q
    enabled    = true
  }
}
`, name, description, projectID)
}

// testAccGCPIntegrationConfig_withFeatures returns a GCP integration with features
func testAccGCPIntegrationConfig_withFeatures(name, description string) string {
	return fmt.Sprintf(`
resource "hush_gcp_integration" "test" {
  name        = %[1]q
  description = %[2]q

  feature {
    name    = "iam"
    enabled = true
  }

  feature {
    name    = "secret_manager"
    enabled = true
  }
}
`, name, description)
}

// testAccGCPIntegrationConfig_full returns a full GCP integration with projects and features
func testAccGCPIntegrationConfig_full(name, description, projectID string) string {
	return fmt.Sprintf(`
resource "hush_gcp_integration" "test" {
  name        = %[1]q
  description = %[2]q

  project {
    project_id = %[3]q
    enabled    = true
  }

  feature {
    name    = "iam"
    enabled = true
  }
}
`, name, description, projectID)
}

// testAccGCPIntegrationConfig_fullUpdated returns an updated full GCP integration
func testAccGCPIntegrationConfig_fullUpdated(name, description, projectID string) string {
	return fmt.Sprintf(`
resource "hush_gcp_integration" "test" {
  name        = %[1]q
  description = %[2]q

  project {
    project_id = %[3]q
    enabled    = true
  }

  feature {
    name    = "iam"
    enabled = true
  }

  feature {
    name    = "secret_manager"
    enabled = true
  }
}
`, name, description, projectID)
}

// testAccGCPIntegrationDataSourceConfig_byID returns a data source configuration looking up by ID
const testAccGCPIntegrationDataSourceConfig_byID = `
data "hush_gcp_integration" "test" {
  id = hush_gcp_integration.test.id
}
`

// testAccGCPIntegrationDataSourceConfig_byName returns a data source configuration looking up by name
const testAccGCPIntegrationDataSourceConfig_byName = `
data "hush_gcp_integration" "test" {
  name = hush_gcp_integration.test.name
}
`
