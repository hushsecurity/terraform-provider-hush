package access_policy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Access policy resource for managing Hush Security access policies"

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: resourceAccessPolicyCreate,
		ReadContext:   accessPolicyRead,
		UpdateContext: resourceAccessPolicyUpdate,
		DeleteContext: resourceAccessPolicyDelete,

		Schema: AccessPolicyResourceSchema(),
	}
}

func resourceAccessPolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateAccessPolicyInput{
		Name:                d.Get("name").(string),
		Enabled:             d.Get("enabled").(bool),
		AccessCredentialID:  d.Get("access_credential_id").(string),
		DeploymentIDs:       expandStringList(d.Get("deployment_ids").([]any)),
		AttestationCriteria: expandAttestationCriteria(d.Get("attestation_criteria").([]any)),
		DeliveryConfig:      expandDeliveryConfig(d),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	if v, ok := d.GetOk("access_privilege_ids"); ok {
		input.AccessPrivilegeIDs = expandStringList(v.([]any))
	}

	policy, err := client.CreateAccessPolicy(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.ID)

	return accessPolicyRead(ctx, d, meta)
}

func resourceAccessPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.UpdateAccessPolicyInput{}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		input.Description = &description
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		input.Enabled = &enabled
	}

	if d.HasChange("access_credential_id") {
		credID := d.Get("access_credential_id").(string)
		input.AccessCredentialID = &credID
	}

	if d.HasChange("access_privilege_ids") {
		privilegeIDs := expandStringList(d.Get("access_privilege_ids").([]any))
		input.AccessPrivilegeIDs = &privilegeIDs
	}

	if d.HasChange("deployment_ids") {
		deploymentIDs := expandStringList(d.Get("deployment_ids").([]any))
		input.DeploymentIDs = &deploymentIDs
	}

	if d.HasChange("attestation_criteria") {
		criteria := expandAttestationCriteria(d.Get("attestation_criteria").([]any))
		input.AttestationCriteria = &criteria
	}

	if d.HasChange("env_delivery_config") {
		deliveryConfig := expandDeliveryConfig(d)
		input.DeliveryConfig = &deliveryConfig
	}

	_, err := client.UpdateAccessPolicy(ctx, c, d.Id(), input)
	if err != nil {
		return diag.FromErr(err)
	}

	return accessPolicyRead(ctx, d, meta)
}

func resourceAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	err := client.DeleteAccessPolicy(ctx, c, d.Id())
	if err != nil && !isNotFoundError(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
