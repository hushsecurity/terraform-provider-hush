package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/deployment"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/notification_channel"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/notification_configuration"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key_id": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("HUSH_API_KEY_ID", nil),
				},
				"api_key_secret": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("HUSH_API_KEY_SECRET", nil),
				},
				"realm": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "US",
					DefaultFunc:  schema.EnvDefaultFunc("HUSH_REALM", "US"),
					Description:  "The Hush realm",
					ValidateFunc: validation.StringInSlice([]string{"US", "EU"}, false),
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"hush_deployment":                 deployment.Resource(),
				"hush_notification_channel":       notification_channel.Resource(),
				"hush_notification_configuration": notification_configuration.Resource(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"hush_deployment":                 deployment.DataSource(),
				"hush_notification_channel":       notification_channel.DataSource(),
				"hush_notification_configuration": notification_configuration.DataSource(),
			},
		}
		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-hush", version)

		apiKeyID := d.Get("api_key_id").(string)
		apiKeySecret := d.Get("api_key_secret").(string)
		realm := strings.ToLower(d.Get("realm").(string))

		var baseURL string

		// Check for development override (for internal development only)
		if devURL := os.Getenv("HUSH_DEV_BASE_URL"); devURL != "" {
			baseURL = devURL
		} else {
			// Production realm mapping - build URL dynamically
			baseURL = fmt.Sprintf("https://api.%s.hush-security.com", realm)
		}

		// TODO: Update client.NewClient to accept userAgent parameter
		_ = userAgent // Suppress unused variable warning until client supports it
		c, err := client.NewClient(ctx, apiKeyID, apiKeySecret, baseURL)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return c, nil
	}
}
