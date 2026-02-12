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
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/access_policy"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/deployment"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/kv_access_credential"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/mongodb_access_credential"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/notification_channel"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/notification_configuration"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/plaintext_access_credential"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/postgres_access_credential"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/postgres_access_privilege"
)

const (
	envHushAPIKeyID     = "HUSH_API_KEY_ID"
	envHushAPIKeySecret = "HUSH_API_KEY_SECRET"
	envHushRealm        = "HUSH_REALM"
	envHushDevBaseURL   = "HUSH_DEV_BASE_URL"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key_id": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc(envHushAPIKeyID, nil),
				},
				"api_key_secret": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc(envHushAPIKeySecret, nil),
				},
				"realm": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "US",
					DefaultFunc:  schema.EnvDefaultFunc(envHushRealm, "US"),
					Description:  "The Hush realm",
					ValidateFunc: validation.StringInSlice([]string{"US", "EU"}, false),
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"hush_deployment":                  deployment.Resource(),
				"hush_notification_channel":        notification_channel.Resource(),
				"hush_notification_configuration":  notification_configuration.Resource(),
				"hush_plaintext_access_credential": plaintext_access_credential.Resource(),
				"hush_kv_access_credential":        kv_access_credential.Resource(),
				"hush_access_policy":               access_policy.Resource(),
				"hush_postgres_access_credential":  postgres_access_credential.Resource(),
				"hush_postgres_access_privilege":   postgres_access_privilege.Resource(),
				"hush_mongodb_access_credential":   mongodb_access_credential.Resource(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"hush_deployment":                  deployment.DataSource(),
				"hush_notification_channel":        notification_channel.DataSource(),
				"hush_notification_configuration":  notification_configuration.DataSource(),
				"hush_plaintext_access_credential": plaintext_access_credential.DataSource(),
				"hush_kv_access_credential":        kv_access_credential.DataSource(),
				"hush_access_policy":               access_policy.DataSource(),
				"hush_postgres_access_credential":  postgres_access_credential.DataSource(),
				"hush_postgres_access_privilege":   postgres_access_privilege.DataSource(),
				"hush_mongodb_access_credential":   mongodb_access_credential.DataSource(),
			},
		}
		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-hush", version)

		apiKeyID := d.Get("api_key_id").(string)
		apiKeySecret := d.Get("api_key_secret").(string)
		realm := strings.ToLower(d.Get("realm").(string))

		var baseURL string

		// Check for development override (for internal development only)
		if devURL := os.Getenv(envHushDevBaseURL); devURL != "" {
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
