package service_agent

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/dsschema"
)

const (
	dataSourceDescription = "A service agent, read by id or by name. Its client secrets are not shown."
	dataIDDesc            = "The agent's id. Exactly one of `id` and `name` is required."
	dataNameDesc          = "The agent's name. Exactly one of `id` and `name` is required; " +
		"it must match a single agent."
)

func DataSource() *schema.Resource {
	s := dsschema.FromResource(ResourceSchema(), dsschema.Options{})
	s["id"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	dsschema.Lookup(s, "id", dataIDDesc, "name")
	dsschema.Lookup(s, "name", dataNameDesc, "id")
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: agentDataRead,
		Schema:      s,
	}
}

func agentDataRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var agent *client.ServiceAgent
	id, name := d.Get("id").(string), d.Get("name").(string)
	if id == "" && name == "" {
		return diag.Errorf("one of id or name is required")
	}
	if id != "" {
		found, err := client.GetServiceAgent(ctx, c, id)
		if err != nil {
			if client.IsNotFoundError(err) {
				return diag.Errorf("no service agent found with id: %s", id)
			}
			return diag.FromErr(err)
		}
		agent = found
	} else {
		agents, err := client.ListServiceAgents(ctx, c)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range agents {
			if agents[i].Name != name {
				continue
			}
			if agent != nil {
				return diag.Errorf("several service agents are named %q; read one by id", name)
			}
			agent = &agents[i]
		}
		if agent == nil {
			return diag.Errorf("no service agent found with name: %s", name)
		}
	}
	d.SetId(agent.ID)
	if err := flatten(d, agent); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
