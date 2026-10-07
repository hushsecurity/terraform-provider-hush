package service_agent_client_secret

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	resourceDescription = "A client secret for a service agent, for workloads that have no identity " +
		"to sign in with. Hush generates it and returns it only when it is created, so it is kept in " +
		"Terraform state. An agent can have two at a time, which is how to rotate: add a second " +
		"resource, move the workload to its secret, then remove the first."
	agentIDDesc = "The service agent's id."
	secretDesc  = "The secret, sent as the client secret with Basic auth (the agent's id as the " +
		"client id). Known only for a secret created by this resource: an imported one is empty."
	hintDesc = "The secret's last 4 characters, to tell secrets apart."
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: secretCreate,
		ReadContext:   secretRead,
		DeleteContext: secretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: secretImport,
		},
		Schema: map[string]*schema.Schema{
			"agent_id": {
				Description:  agentIDDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(idPart, "must be an agent id"),
			},
			"secret": {Description: secretDesc, Type: schema.TypeString, Computed: true, Sensitive: true},
			"hint":   {Description: hintDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

var idPart = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// The id is "<agent id>/<credential id>": a secret is reached through its agent.
func parseID(id string) (agentID, credentialID string, err error) {
	agentID, credentialID, ok := strings.Cut(id, "/")
	if !ok || !idPart.MatchString(agentID) || !idPart.MatchString(credentialID) {
		return "", "", fmt.Errorf("expected <agent id>/<secret id>, got %q", id)
	}
	return agentID, credentialID, nil
}

func secretCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	agentID := d.Get("agent_id").(string)

	created, err := client.AddAgentCredential(ctx, c, agentID,
		map[string]any{"type": client.CredentialClientSecret})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(agentID + "/" + created.ID)
	if err := d.Set("secret", created.Secret); err != nil {
		return diag.FromErr(err)
	}
	return secretRead(ctx, d, m)
}

func secretRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	agentID, credentialID, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	agent, err := client.GetServiceAgent(ctx, c, agentID)
	if err != nil {
		if client.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	for _, credential := range agent.Credentials {
		if credential.ID == credentialID && credential.Type == client.CredentialClientSecret {
			if err := d.Set("agent_id", agentID); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("hint", credential.Hint); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}
	d.SetId("")
	return nil
}

func secretDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	agentID, credentialID, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.RemoveAgentCredential(ctx, c, agentID, credentialID); err != nil && !client.IsNotFoundError(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func secretImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	agentID, _, err := parseID(d.Id())
	if err != nil {
		return nil, err
	}
	if err := d.Set("agent_id", agentID); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
