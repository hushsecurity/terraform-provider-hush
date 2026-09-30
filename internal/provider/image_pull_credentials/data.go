package image_pull_credentials

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	dataSourceDescription = "The container registry credentials for pulling the Hush container images. " +
		"They are issued per organization rather than per deployment, so the data source takes no " +
		"arguments and every `hush_deployment` pulls with the same ones."
	usernameDesc = "The username to authenticate against the registry with."
	passwordDesc = "The password to authenticate against the registry with."
	registryDesc = "The registry the Hush container images are hosted on."
	tokenDesc    = "The registry, username and password in one opaque bundle, which the installers accept in place of the three."

	// The credentials are an organization-wide singleton, so there is nothing
	// to look them up by and no id of their own to report.
	singletonID = "image_pull_credentials"
)

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: imagePullCredentialsRead,
		Schema: map[string]*schema.Schema{
			"username": {Description: usernameDesc, Type: schema.TypeString, Computed: true, Sensitive: true},
			"password": {Description: passwordDesc, Type: schema.TypeString, Computed: true, Sensitive: true},
			"registry": {Description: registryDesc, Type: schema.TypeString, Computed: true},
			"token":    {Description: tokenDesc, Type: schema.TypeString, Computed: true, Sensitive: true},
		},
	}
}

func imagePullCredentialsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	creds, err := client.GetImagePullCredentials(ctx, c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read image pull credentials: %w", err))
	}

	d.SetId(singletonID)
	for key, value := range map[string]any{
		"username": creds.Username,
		"password": creds.Password,
		"registry": creds.Registry,
		"token":    creds.Token,
	} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", key, err))
		}
	}
	return nil
}
