package notification_channel

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
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
			"hush_notification_channel": Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hush_notification_channel": DataSource(),
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

func TestAccResourceNotificationChannelEmail_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckNotificationChannelDestroy,
		Steps: []resource.TestStep{
			// Create step
			{
				Config: testAccNotificationChannelConfig_email("email-channel", "email channel description", []string{"user1@example.com", "user2@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists("hush_notification_channel.email"),
					resource.TestMatchResourceAttr(
						"hush_notification_channel.email", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "name", "email-channel",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "description", "email channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.0.recipients.0", "user1@example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.0.recipients.1", "user2@example.com",
					),
				),
			},
			// Update step
			{
				Config: testAccNotificationChannelConfig_email("email-channel-updated", "updated email channel description", []string{"user3@example.com"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists("hush_notification_channel.email"),
					resource.TestMatchResourceAttr(
						"hush_notification_channel.email", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "name", "email-channel-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "description", "updated email channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.0.recipients.0", "user3@example.com",
					),
				),
			},
			// Import step
			{
				ResourceName:      "hush_notification_channel.email",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNotificationChannelWebhook_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckNotificationChannelDestroy,
		Steps: []resource.TestStep{
			// Create step
			{
				Config: testAccNotificationChannelConfig_webhook("webhook-channel", "webhook channel description", "https://example.com/webhook", "POST", map[string]string{"Content-Type": "application/json"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists("hush_notification_channel.webhook"),
					resource.TestMatchResourceAttr(
						"hush_notification_channel.webhook", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "name", "webhook-channel",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "description", "webhook channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.url", "https://example.com/webhook",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.method", "POST",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.headers.Content-Type", "application/json",
					),
				),
			},
			// Update step
			{
				Config: testAccNotificationChannelConfig_webhook("webhook-channel-updated", "updated webhook channel description", "https://api.example.com/notifications", "POST", map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists("hush_notification_channel.webhook"),
					resource.TestMatchResourceAttr(
						"hush_notification_channel.webhook", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "name", "webhook-channel-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "description", "updated webhook channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.url", "https://api.example.com/notifications",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.headers.Authorization", "Bearer token",
					),
				),
			},
			// Import step
			{
				ResourceName:      "hush_notification_channel.webhook",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNotificationChannelSlack_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckNotificationChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_slack("slack-channel", "slack channel description", "int-euIk8SVlvEGqOzNM5D", "test-channel"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNotificationChannelExists("hush_notification_channel.slack"),
					resource.TestMatchResourceAttr(
						"hush_notification_channel.slack", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "name", "slack-channel",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "description", "slack channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "slack_config.0.integration_id", "int-euIk8SVlvEGqOzNM5D",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "slack_config.0.channel_name", "test-channel",
					),
				),
			},
		},
	})
}

func TestAccDataSourceNotificationChannel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckNotificationChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelConfig_email("email-channel", "email channel description", []string{"user1@example.com", "user2@example.com"}) + testAccNotificationChannelDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_notification_channel.channel", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "name", "email-channel",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "description", "email channel description",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "email_config.0.recipients.0", "user1@example.com",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "email_config.0.recipients.1", "user2@example.com",
					),
				),
			},
		},
	})
}

// testAccCheckNotificationChannelDestroy verifies all notification channel resources have been destroyed
func testAccCheckNotificationChannelDestroy(s *terraform.State) error {
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
		if rs.Type != "hush_notification_channel" {
			continue
		}

		_, err := client.GetNotificationChannel(context.Background(), c, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("notification channel %s still exists", rs.Primary.ID)
		}

		apiError, ok := err.(*client.APIError)
		if ok && apiError.IsNotFound() {
			continue // Resource properly destroyed
		}
		return fmt.Errorf("failed to verify notification channel %s was destroyed: %s", rs.Primary.ID, err)
	}
	return nil
}

// testAccCheckNotificationChannelExists verifies a notification channel resource exists in the API
func testAccCheckNotificationChannelExists(resourceName string) resource.TestCheckFunc {
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

		_, err = client.GetNotificationChannel(context.Background(), c, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("notification channel %s not found: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

// testAccNotificationChannelConfig_email returns an email notification channel configuration
func testAccNotificationChannelConfig_email(name, description string, recipients []string) string {
	recipientsStr := ""
	for i, recipient := range recipients {
		if i > 0 {
			recipientsStr += ", "
		}
		recipientsStr += fmt.Sprintf("%q", recipient)
	}

	return fmt.Sprintf(`
resource "hush_notification_channel" "email" {
  name        = %[1]q
  description = %[2]q
  email_config {
    recipients = [%[3]s]
  }
}
`, name, description, recipientsStr)
}

// testAccNotificationChannelConfig_webhook returns a webhook notification channel configuration
func testAccNotificationChannelConfig_webhook(name, description, url, method string, headers map[string]string) string {
	var headersLines []string
	for key, value := range headers {
		headersLines = append(headersLines, fmt.Sprintf("    %q = %q", key, value))
	}
	headersStr := strings.Join(headersLines, "\n")

	return fmt.Sprintf(`
resource "hush_notification_channel" "webhook" {
  name        = %[1]q
  description = %[2]q
  webhook_config {
    url    = %[3]q
    method = %[4]q
    headers = {
%[5]s
    }
  }
}
`, name, description, url, method, headersStr)
}

// testAccNotificationChannelConfig_slack returns a slack notification channel configuration
func testAccNotificationChannelConfig_slack(name, description, integrationId, channelName string) string {
	return fmt.Sprintf(`
resource "hush_notification_channel" "slack" {
  name        = %[1]q
  description = %[2]q
  slack_config {
    integration_id = %[3]q
    channel_name   = %[4]q
  }
}
`, name, description, integrationId, channelName)
}

// testAccNotificationChannelDataSourceConfig returns a data source configuration
const testAccNotificationChannelDataSourceConfig = `
data "hush_notification_channel" "channel" {
  id = hush_notification_channel.email.id
}
`
