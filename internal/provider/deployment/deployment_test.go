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

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
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

const (
	oidcTestIssuer   = "https://oidc.eks.us-east-1.amazonaws.com/id/D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9"
	oidcTestAudience = "https://kubernetes.default.svc"
)

// TestAccResourceDeployment_oidc exercises the full oidc_provider lifecycle:
// create with the block, update it, remove it (explicit null to the API), and
// import.
func TestAccResourceDeployment_oidc(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckDeploymentDestroy,
		Steps: []resource.TestStep{
			// Create with oidc_provider
			{
				Config: testAccDeploymentConfig_oidc("test-oidc", oidcTestAudience, `["system:serviceaccount:hush-security:hush-agent"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists("hush_deployment.test"),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.issuer", oidcTestIssuer),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.audience", oidcTestAudience),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.allowed_subjects.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.allowed_subjects.0", "system:serviceaccount:hush-security:hush-agent"),
				),
			},
			// Update the block (widen allowed subjects)
			{
				Config: testAccDeploymentConfig_oidc("test-oidc", oidcTestAudience, `["system:serviceaccount:hush-security:*", "system:serviceaccount:other:*"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.allowed_subjects.#", "2"),
				),
			},
			// Remove the block
			{
				Config: testAccDeploymentConfig_noOidc("test-oidc"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.#", "0"),
				),
			},
			// Import
			{
				ResourceName:      "hush_deployment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccDeploymentConfig_oidc returns a deployment configuration with an
// oidc_provider block. allowedSubjects is injected verbatim as an HCL list.
func testAccDeploymentConfig_oidc(name, audience, allowedSubjects string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = %[1]q
  kind = "k8s"

  oidc_provider {
    issuer           = %[2]q
    audience         = %[3]q
    allowed_subjects = %[4]s
  }
}
`, name, oidcTestIssuer, audience, allowedSubjects)
}

// testAccDeploymentConfig_noOidc returns a deployment configuration without an
// oidc_provider block, used to verify removal.
func testAccDeploymentConfig_noOidc(name string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = %[1]q
  kind = "k8s"
}
`, name)
}
