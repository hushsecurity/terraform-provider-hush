package access_policy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc                  = "The ID of the access policy"
	nameDesc                = "The name of the access policy"
	descriptionDesc         = "The description of the access policy"
	enabledDesc             = "Whether the access policy is enabled"
	accessCredentialIDDesc  = "The ID of the access credential"
	accessPrivilegeIDsDesc  = "The list of access privilege IDs"
	deploymentIDsDesc       = "The list of deployment IDs"
	attestationCriteriaDesc = "The attestation criteria for the access policy"
	envDeliveryConfigDesc   = "Environment variable delivery configuration for the access policy"
	statusDesc              = "The status of the access policy (syncing, ok, warning, error, disabled)"
	statusDetailDesc        = "The status detail of the access policy"
)

func AccessPolicyResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: idDesc,
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  nameDesc,
			ValidateFunc: validation.StringLenBetween(1, 255),
		},
		"description": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  descriptionDesc,
			ValidateFunc: validation.StringLenBetween(0, 1000),
		},
		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: enabledDesc,
		},
		"access_credential_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: accessCredentialIDDesc,
		},
		"access_privilege_ids": {
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: accessPrivilegeIDsDesc,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"deployment_ids": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: deploymentIDsDesc,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
			},
		},
		"attestation_criteria": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: attestationCriteriaDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"k8s:ns", "k8s:sa", "k8s:pod-label", "k8s:pod-name", "k8s:container-name"}, false),
						Description:  "The type of attestation criterion (k8s:ns, k8s:sa, k8s:pod-label, k8s:pod-name, or k8s:container-name)",
					},
					"value": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The value of the attestation criterion",
					},
					"key": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The key for label type criterion",
					},
				},
			},
		},
		"env_delivery_config": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: envDeliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"item": {
						Type:        schema.TypeList,
						Required:    true,
						MinItems:    1,
						Description: "The delivery items",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The credential key or template string for the delivery item",
								},
								"name": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The environment variable name for the delivery item",
								},
								"type": {
									Type:         schema.TypeString,
									Optional:     true,
									Default:      string(client.EnvMappingTypeKey),
									ValidateFunc: validation.StringInSlice([]string{string(client.EnvMappingTypeKey), string(client.EnvMappingTypeTemplate)}, false),
									Description:  "The type of delivery item mapping (key or template)",
								},
							},
						},
					},
				},
			},
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: statusDesc,
		},
		"status_detail": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: statusDetailDesc,
		},
	}
}

func AccessPolicyDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: idDesc,
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: nameDesc,
		},
		"description": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: descriptionDesc,
		},
		"enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: enabledDesc,
		},
		"access_credential_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: accessCredentialIDDesc,
		},
		"access_privilege_ids": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: accessPrivilegeIDsDesc,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"deployment_ids": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: deploymentIDsDesc,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"attestation_criteria": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: attestationCriteriaDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The type of attestation criterion (k8s:ns, k8s:sa, k8s:pod-label, k8s:pod-name, or k8s:container-name)",
					},
					"value": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The value of the attestation criterion",
					},
					"key": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The key for label type criterion",
					},
				},
			},
		},
		"env_delivery_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: envDeliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"item": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The delivery items",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The credential key or template string for the delivery item",
								},
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The environment variable name for the delivery item",
								},
								"type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The type of delivery item mapping (key or template)",
								},
							},
						},
					},
				},
			},
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: statusDesc,
		},
		"status_detail": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: statusDetailDesc,
		},
	}
}

// Helper Functions

func accessPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	id := d.Id()
	if id == "" {
		if v, ok := d.GetOk("id"); ok {
			id = v.(string)
		}
	}

	policy, err := client.GetAccessPolicy(ctx, c, id)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(policy.ID)

	if diags := setAccessPolicyFields(d, policy); diags.HasError() {
		return diags
	}

	return nil
}

func setAccessPolicyFields(d *schema.ResourceData, policy *client.AccessPolicy) diag.Diagnostics {
	fields := map[string]any{
		"name":                 policy.Name,
		"description":          policy.Description,
		"enabled":              policy.Enabled,
		"access_credential_id": policy.AccessCredentialID,
		"access_privilege_ids": policy.AccessPrivilegeIDs,
		"deployment_ids":       policy.DeploymentIDs,
		"attestation_criteria": flattenAttestationCriteria(policy.AttestationCriteria),
		"status":               policy.Status,
		"status_detail":        policy.StatusDetail,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if diags := flattenDeliveryConfig(d, policy.DeliveryConfig); diags.HasError() {
		return diags
	}

	return nil
}

func expandStringList(list []any) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

func expandAttestationCriteria(list []any) []client.AttestationCriterion {
	result := make([]client.AttestationCriterion, len(list))
	for i, v := range list {
		m := v.(map[string]any)
		criterion := client.AttestationCriterion{
			Type:  client.AttestationCriterionType(m["type"].(string)),
			Value: m["value"].(string),
		}
		if key, ok := m["key"].(string); ok && key != "" {
			criterion.Key = key
		}
		result[i] = criterion
	}
	return result
}

func expandDeliveryConfig(d *schema.ResourceData) client.DeliveryConfig {
	if v, ok := d.GetOk("env_delivery_config"); ok {
		return expandEnvDeliveryConfig(v.([]any))
	}

	return client.DeliveryConfig{}
}

func expandEnvDeliveryConfig(list []any) client.DeliveryConfig {
	if len(list) == 0 || list[0] == nil {
		return client.DeliveryConfig{}
	}

	m := list[0].(map[string]any)
	config := client.DeliveryConfig{
		Type: client.DeliveryTypeEnv,
	}

	if items, ok := m["item"].([]any); ok {
		config.Items = make([]any, len(items))
		for i, item := range items {
			itemMap := item.(map[string]any)
			deliveryItem := client.EnvDeliveryItem{
				Name: itemMap["name"].(string), // Environment variable name
			}
			if key, ok := itemMap["key"].(string); ok && key != "" {
				deliveryItem.Key = key
			}
			if t, ok := itemMap["type"].(string); ok && t != "" {
				deliveryItem.Type = client.EnvMappingType(t)
			}
			config.Items[i] = deliveryItem
		}
	}

	return config
}

func flattenAttestationCriteria(criteria []client.AttestationCriterion) []any {
	result := make([]any, len(criteria))
	for i, c := range criteria {
		m := map[string]any{
			"type":  string(c.Type),
			"value": c.Value,
		}
		if c.Key != "" {
			m["key"] = c.Key
		}
		result[i] = m
	}
	return result
}

func flattenDeliveryConfig(d *schema.ResourceData, config client.DeliveryConfig) diag.Diagnostics {
	switch config.Type {
	case client.DeliveryTypeEnv:
		if err := d.Set("env_delivery_config", []any{map[string]any{"item": config.Items}}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set env_delivery_config: %w", err))
		}
	}

	return nil
}

func isNotFoundError(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}
