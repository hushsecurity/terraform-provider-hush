package service_agent

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: agentCreate,
		ReadContext:   agentRead,
		UpdateContext: agentUpdate,
		DeleteContext: agentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			return validateCredentials(d)
		},
		Schema: ResourceSchema(),
	}
}

func agentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	agent, err := client.CreateServiceAgent(ctx, c, &client.ServiceAgentInput{
		Name:         d.Get("name").(string),
		Enabled:      d.Get("enabled").(bool),
		ActsForUsers: d.Get("acts_for_users").(bool),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	// Set before the credentials: if one is refused, the agent is still in
	// state, tainted, rather than lost.
	d.SetId(agent.ID)
	if err := syncCredentials(ctx, c, d, nil); err != nil {
		return diag.FromErr(err)
	}
	return agentRead(ctx, d, m)
}

func agentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	agent, err := client.GetServiceAgent(ctx, c, d.Id())
	if err != nil {
		if client.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := flatten(d, agent); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func agentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	patch := &client.ServiceAgentUpdate{}
	changed := false
	if d.HasChange("name") {
		name := d.Get("name").(string)
		patch.Name = &name
		changed = true
	}
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		patch.Enabled = &enabled
		changed = true
	}
	if d.HasChange("acts_for_users") {
		acts := d.Get("acts_for_users").(bool)
		patch.ActsForUsers = &acts
		changed = true
	}
	var agent *client.ServiceAgent
	if changed {
		updated, err := client.UpdateServiceAgent(ctx, c, d.Id(), patch)
		if err != nil {
			return diag.FromErr(err)
		}
		agent = updated
	}
	if d.HasChanges(client.FederatedCredentialTypes...) {
		if agent == nil {
			fetched, err := client.GetServiceAgent(ctx, c, d.Id())
			if err != nil {
				return diag.FromErr(err)
			}
			agent = fetched
		}
		if err := syncCredentials(ctx, c, d, agent.Credentials); err != nil {
			return diag.FromErr(err)
		}
	}
	return agentRead(ctx, d, m)
}

// syncCredentials makes the agent's federated credentials match its blocks.
// Credentials can't be changed, so a changed block is a removal and an add.
// New ones go in before old ones come out, so a failed apply never leaves the
// agent with fewer identities than it had. Old ones go first only for a
// replacement claiming the same identity (heimdall binds one once), or to
// make room when the agent is at heimdall's limit.
func syncCredentials(ctx context.Context, c *client.Client, d *schema.ResourceData, existing []client.AgentCredential) error {
	wanted, err := configuredCredentials(d)
	if err != nil {
		return err
	}
	wantedKeys := map[string]bool{}
	for _, w := range wanted {
		wantedKeys[w.key] = true
	}
	heldKeys := map[string]bool{}
	staleByIdentity := map[string]credential{}
	var stale []credential
	held := 0
	for _, e := range existing {
		if !isFederated(e.Type) {
			continue
		}
		have, err := existingCredential(e)
		if err != nil {
			return err
		}
		held++
		if wantedKeys[have.key] {
			heldKeys[have.key] = true
			continue
		}
		stale = append(stale, have)
		staleByIdentity[have.identity] = have
	}

	removed := map[string]bool{}
	remove := func(old credential) error {
		if removed[old.id] {
			return nil
		}
		if err := client.RemoveAgentCredential(ctx, c, d.Id(), old.id); err != nil && !client.IsNotFoundError(err) {
			return err
		}
		removed[old.id] = true
		held--
		return nil
	}
	for _, w := range wanted {
		if heldKeys[w.key] {
			continue
		}
		if old, ok := staleByIdentity[w.identity]; ok {
			if err := remove(old); err != nil {
				return err
			}
		}
		for held >= client.MaxAgentCredentials {
			next := firstNotRemoved(stale, removed)
			if next == nil {
				break // the plan refuses more than the limit; let heimdall say why
			}
			if err := remove(*next); err != nil {
				return err
			}
		}
		if _, err := client.AddAgentCredential(ctx, c, d.Id(), w.body); err != nil {
			return err
		}
		held++
	}
	for _, old := range stale {
		if err := remove(old); err != nil {
			return err
		}
	}
	return nil
}

func firstNotRemoved(stale []credential, removed map[string]bool) *credential {
	for i := range stale {
		if !removed[stale[i].id] {
			return &stale[i]
		}
	}
	return nil
}

func agentDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if err := client.DeleteServiceAgent(ctx, c, d.Id()); err != nil && !client.IsNotFoundError(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
