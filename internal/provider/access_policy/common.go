package access_policy

import (
	"context"
	"fmt"

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
	deploymentIDsDesc       = "The list of deployment IDs"
	attestationCriteriaDesc = "The attestation criteria for the access policy"
	deliveryConfigDesc      = "The delivery configuration for the access policy"
	createdAtDesc           = "The creation timestamp"
	modifiedAtDesc          = "The modification timestamp"
)

func AccessPolicyResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: idDesc,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: nameDesc,
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: descriptionDesc,
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
		"deployment_ids": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: deploymentIDsDesc,
			Elem: &schema.Schema{
				Type: schema.TypeString,
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
		"delivery_config": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: deliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"env"}, false),
						Description:  "The type of delivery (only 'env' is supported)",
					},
					"items": {
						Type:        schema.TypeList,
						Required:    true,
						MinItems:    1,
						Description: "The delivery items",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The key for the delivery item",
								},
								"name": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The environment variable name for the delivery item",
								},
							},
						},
					},
				},
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: createdAtDesc,
		},
		"modified_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: modifiedAtDesc,
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
						Description: "The type of attestation criterion (namespace or label)",
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
		"delivery_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: deliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The type of delivery (env or file)",
					},
					"items": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The delivery items",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The key for the delivery item",
								},
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The environment variable name for the delivery item",
								},
							},
						},
					},
				},
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: createdAtDesc,
		},
		"modified_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: modifiedAtDesc,
		},
	}
}

// Helper Functions

func accessPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	fields := map[string]interface{}{
		"name":                 policy.Name,
		"description":          policy.Description,
		"enabled":              policy.Enabled,
		"access_credential_id": policy.AccessCredentialID,
		"deployment_ids":       policy.DeploymentIDs,
		"attestation_criteria": flattenAttestationCriteria(policy.AttestationCriteria),
		"delivery_config":      flattenDeliveryConfig(policy.DeliveryConfig),
		"created_at":           policy.CreatedAt,
		"modified_at":          policy.ModifiedAt,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}

func expandStringList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

func expandAttestationCriteria(list []interface{}) []client.AttestationCriterion {
	result := make([]client.AttestationCriterion, len(list))
	for i, v := range list {
		m := v.(map[string]interface{})
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

func expandDeliveryConfig(list []interface{}) client.DeliveryConfig {
	if len(list) == 0 || list[0] == nil {
		return client.DeliveryConfig{}
	}

	m := list[0].(map[string]interface{})
	config := client.DeliveryConfig{
		Type: client.DeliveryType(m["type"].(string)),
	}

	if items, ok := m["items"].([]interface{}); ok {
		config.Items = make([]client.DeliveryItem, len(items))
		for i, item := range items {
			itemMap := item.(map[string]interface{})
			config.Items[i] = client.DeliveryItem{
				Key:  itemMap["key"].(string),  // Credential key field
				Name: itemMap["name"].(string), // Environment variable name
			}
		}
	}

	return config
}

func flattenAttestationCriteria(criteria []client.AttestationCriterion) []interface{} {
	result := make([]interface{}, len(criteria))
	for i, c := range criteria {
		m := map[string]interface{}{
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

func flattenDeliveryConfig(config client.DeliveryConfig) []interface{} {
	items := make([]interface{}, len(config.Items))
	for i, item := range config.Items {
		items[i] = map[string]interface{}{
			"key":  item.Key,
			"name": item.Name,
		}
	}

	return []interface{}{
		map[string]interface{}{
			"type":  string(config.Type),
			"items": items,
		},
	}
}

func isNotFoundError(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}
