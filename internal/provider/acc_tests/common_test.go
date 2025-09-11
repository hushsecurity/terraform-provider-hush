package acc_tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	p "github.com/hushsecurity/terraform-provider-hush/internal/provider"
)

var provider *schema.Provider

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"hush": func() (*schema.Provider, error) {
		if provider == nil {
			provider = p.New("dev")()
		}
		return provider, nil
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

func validateResourceDestroyed(resource, resourcePath string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		c := provider.Meta().(*client.Client)
		resourceType := fmt.Sprintf("hush_%s", resource)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			resourceId := rs.Primary.ID

			var err error
			switch resource {
			case "notification_channel":
				_, err = client.GetNotificationChannel(context.Background(), c, resourceId)
			case "notification_configuration":
				_, err = client.GetNotificationConfiguration(context.Background(), c, resourceId)
			case "deployment":
				_, err = client.GetDeployment(context.Background(), c, resourceId)
			default:
				return fmt.Errorf("unknown resource type: %s", resource)
			}

			if err == nil {
				return fmt.Errorf("%s %s still exists", resource, resourceId)
			}
			apiError, ok := err.(*client.APIError)
			if ok && apiError.IsNotFound() {
				return nil
			}
			return fmt.Errorf("failed to verify %s %s was destroyed: %s", resource, resourceId, err)
		}
		return nil
	}
}
