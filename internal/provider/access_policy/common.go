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
	idDesc                   = "The ID of the access policy"
	nameDesc                 = "The name of the access policy"
	descriptionDesc          = "The description of the access policy"
	enabledDesc              = "Whether the access policy is enabled"
	accessCredentialIDDesc   = "The ID of the access credential"
	accessPrivilegeIDsDesc   = "The list of access privilege IDs"
	deploymentIDsDesc        = "The list of deployment IDs"
	attestationCriteriaDesc  = "The attestation criteria for the access policy"
	envDeliveryConfigDesc    = "Environment variable delivery configuration for the access policy"
	volumeDeliveryConfigDesc = "Volume mount delivery configuration for the access policy"
	awsWifDeliveryConfigDesc = "AWS WIF delivery configuration for the access policy"
	gcpWifDeliveryConfigDesc = "GCP WIF delivery configuration for the access policy"
	statusDesc               = "The status of the access policy (syncing, ok, warning, error, disabled)"
	statusDetailDesc         = "The status detail of the access policy"
)

var deliveryConfigExactlyOneOf = []string{"env_delivery_config", "volume_delivery_config", "aws_wif_delivery_config", "gcp_wif_delivery_config"}

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
			Type:         schema.TypeList,
			Optional:     true,
			Description:  envDeliveryConfigDesc,
			ExactlyOneOf: deliveryConfigExactlyOneOf,
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
						Default:      string(client.DeliveryMappingTypeKey),
						ValidateFunc: validation.StringInSlice([]string{string(client.DeliveryMappingTypeKey), string(client.DeliveryMappingTypeTemplate)}, false),
						Description:  "The type of delivery item mapping (key or template)",
					},
				},
			},
		},
		"volume_delivery_config": {
			Type:         schema.TypeList,
			Optional:     true,
			MaxItems:     1,
			Description:  volumeDeliveryConfigDesc,
			ExactlyOneOf: deliveryConfigExactlyOneOf,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"mount_point": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The absolute path where the volume will be mounted",
					},
					"item": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The relative file path within the mount point",
								},
								"key": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The credential key or template string for the delivery item",
								},
								"type": {
									Type:         schema.TypeString,
									Optional:     true,
									Default:      string(client.DeliveryMappingTypeKey),
									ValidateFunc: validation.StringInSlice([]string{string(client.DeliveryMappingTypeKey), string(client.DeliveryMappingTypeTemplate)}, false),
									Description:  "The type of delivery item mapping (key or template)",
								},
							},
						},
					},
				},
			},
		},
		"aws_wif_delivery_config": {
			Type:         schema.TypeList,
			Optional:     true,
			MaxItems:     1,
			Description:  awsWifDeliveryConfigDesc,
			ExactlyOneOf: deliveryConfigExactlyOneOf,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"role_arn": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The AWS IAM role ARN to assume via WIF",
					},
					"subject_kind": wifSubjectKindResourceSchema(),
					"subject":      wifSubjectResourceSchema(),
				},
			},
		},
		"gcp_wif_delivery_config": {
			Type:         schema.TypeList,
			Optional:     true,
			MaxItems:     1,
			Description:  gcpWifDeliveryConfigDesc,
			ExactlyOneOf: deliveryConfigExactlyOneOf,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"subject_kind": wifSubjectKindResourceSchema(),
					"subject":      wifSubjectResourceSchema(),
					"service_account": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The GCP service account email (e.g. my-sa@my-project.iam.gserviceaccount.com)",
					},
					"service_account_token_lifetime": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3600,
						ValidateFunc: validation.IntAtLeast(1),
						Description:  "The token lifetime in seconds (default: 3600)",
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
		"volume_delivery_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: volumeDeliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"mount_point": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The absolute path where the volume will be mounted",
					},
					"item": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The relative file path within the mount point",
								},
								"key": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The credential key or template string for the delivery item",
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
		"aws_wif_delivery_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: awsWifDeliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"role_arn": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The AWS IAM role ARN to assume via WIF",
					},
					"subject_kind": wifSubjectKindDataSourceSchema(),
					"subject":      wifSubjectDataSourceSchema(),
				},
			},
		},
		"gcp_wif_delivery_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: gcpWifDeliveryConfigDesc,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"subject_kind": wifSubjectKindDataSourceSchema(),
					"subject":      wifSubjectDataSourceSchema(),
					"service_account": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The GCP service account email",
					},
					"service_account_token_lifetime": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "The token lifetime in seconds",
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

func expandDeliveryConfig(d *schema.ResourceData) any {
	if v, ok := d.GetOk("env_delivery_config"); ok {
		return expandEnvDeliveryConfig(v.([]any))
	}
	if v, ok := d.GetOk("volume_delivery_config"); ok {
		return expandVolumeDeliveryConfig(v.([]any))
	}
	if v, ok := d.GetOk("aws_wif_delivery_config"); ok {
		return expandAwsWifDeliveryConfig(v.([]any))
	}
	if v, ok := d.GetOk("gcp_wif_delivery_config"); ok {
		return expandGcpWifDeliveryConfig(v.([]any))
	}

	return nil
}

func expandEnvDeliveryConfig(list []any) *client.EnvDeliveryConfig {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	items := make([]client.EnvDeliveryItem, len(list))
	for i, item := range list {
		itemMap := item.(map[string]any)
		deliveryItem := client.EnvDeliveryItem{
			Name: itemMap["name"].(string),
		}
		if key, ok := itemMap["key"].(string); ok && key != "" {
			deliveryItem.Key = key
		}
		if t, ok := itemMap["type"].(string); ok && t != "" {
			deliveryItem.Type = client.DeliveryMappingType(t)
		}
		items[i] = deliveryItem
	}

	return &client.EnvDeliveryConfig{
		Type:  client.DeliveryTypeEnv,
		Items: items,
	}
}

func expandVolumeDeliveryConfig(list []any) *client.VolumeDeliveryConfig {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	configMap := list[0].(map[string]any)
	rawItems := configMap["item"].([]any)

	items := make([]client.VolumeDeliveryItem, len(rawItems))
	for i, item := range rawItems {
		itemMap := item.(map[string]any)
		deliveryItem := client.VolumeDeliveryItem{
			Path: itemMap["path"].(string),
		}
		if key, ok := itemMap["key"].(string); ok && key != "" {
			deliveryItem.Key = key
		}
		if t, ok := itemMap["type"].(string); ok && t != "" {
			deliveryItem.Type = client.DeliveryMappingType(t)
		}
		items[i] = deliveryItem
	}

	return &client.VolumeDeliveryConfig{
		Type:       client.DeliveryTypeVolume,
		MountPoint: configMap["mount_point"].(string),
		Items:      items,
	}
}

// WIF shared helpers

func wifSubjectKindResourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Default:      string(client.WifSubjectKindHushSubject),
		ValidateFunc: validation.StringInSlice([]string{string(client.WifSubjectKindHushSubject), string(client.WifSubjectKindServiceAccount)}, false),
		Description:  "The subject kind for WIF. hush_subject uses hush:federation:<subject>, service_account uses system:serviceaccount:<namespace>:<serviceaccount>",
	}
}

func wifSubjectResourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The federation subject identifier (required when subject_kind is hush_subject)",
	}
}

func wifSubjectKindDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The subject kind for WIF. hush_subject uses hush:federation:<subject>, service_account uses system:serviceaccount:<namespace>:<serviceaccount>",
	}
}

func wifSubjectDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The federation subject identifier",
	}
}

func expandWifSubject(configMap map[string]any) (client.WifSubjectKind, string) {
	subjectKind := client.WifSubjectKind(configMap["subject_kind"].(string))
	var subject string
	if s, ok := configMap["subject"].(string); ok && s != "" {
		subject = s
	}
	return subjectKind, subject
}

func flattenWifSubject(configMap map[string]any, result map[string]any) {
	result["subject_kind"] = configMap["subject_kind"]
	if subject, ok := configMap["subject"]; ok {
		result["subject"] = subject
	}
}

func expandAwsWifDeliveryConfig(list []any) *client.AwsWifDeliveryConfig {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	configMap := list[0].(map[string]any)
	subjectKind, subject := expandWifSubject(configMap)

	return &client.AwsWifDeliveryConfig{
		Type:        client.DeliveryTypeAwsWif,
		RoleArn:     configMap["role_arn"].(string),
		SubjectKind: subjectKind,
		Subject:     subject,
	}
}

func expandGcpWifDeliveryConfig(list []any) *client.GcpWifDeliveryConfig {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	configMap := list[0].(map[string]any)
	subjectKind, subject := expandWifSubject(configMap)

	config := &client.GcpWifDeliveryConfig{
		Type:                        client.DeliveryTypeGcpWif,
		SubjectKind:                 subjectKind,
		Subject:                     subject,
		ServiceAccountTokenLifetime: configMap["service_account_token_lifetime"].(int),
	}

	if sa, ok := configMap["service_account"].(string); ok && sa != "" {
		config.ServiceAccount = sa
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

func flattenDeliveryConfig(d *schema.ResourceData, config any) diag.Diagnostics {
	configMap, ok := config.(map[string]any)
	if !ok {
		return nil
	}

	deliveryType, _ := configMap["type"].(string)

	switch client.DeliveryType(deliveryType) {
	case client.DeliveryTypeEnv:
		if err := d.Set("env_delivery_config", configMap["items"]); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set env_delivery_config: %w", err))
		}
	case client.DeliveryTypeVolume:
		if err := d.Set("volume_delivery_config", flattenVolumeDeliveryConfig(configMap)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set volume_delivery_config: %w", err))
		}
	case client.DeliveryTypeAwsWif:
		if err := d.Set("aws_wif_delivery_config", flattenAwsWifDeliveryConfig(configMap)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set aws_wif_delivery_config: %w", err))
		}
	case client.DeliveryTypeGcpWif:
		if err := d.Set("gcp_wif_delivery_config", flattenGcpWifDeliveryConfig(configMap)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set gcp_wif_delivery_config: %w", err))
		}
	}

	return nil
}

func flattenVolumeDeliveryConfig(configMap map[string]any) []any {
	rawItems, _ := configMap["items"].([]any)
	items := make([]any, len(rawItems))
	for i, item := range rawItems {
		itemMap, _ := item.(map[string]any)
		items[i] = map[string]any{
			"path": itemMap["path"],
			"key":  itemMap["key"],
			"type": itemMap["type"],
		}
	}

	return []any{
		map[string]any{
			"mount_point": configMap["mount_point"],
			"item":        items,
		},
	}
}

func flattenAwsWifDeliveryConfig(configMap map[string]any) []any {
	result := map[string]any{
		"role_arn": configMap["role_arn"],
	}
	flattenWifSubject(configMap, result)
	return []any{result}
}

func flattenGcpWifDeliveryConfig(configMap map[string]any) []any {
	result := map[string]any{}
	flattenWifSubject(configMap, result)
	if sa, ok := configMap["service_account"]; ok {
		result["service_account"] = sa
	}
	if lifetime, ok := configMap["service_account_token_lifetime"]; ok {
		switch v := lifetime.(type) {
		case float64:
			result["service_account_token_lifetime"] = int(v)
		case int:
			result["service_account_token_lifetime"] = v
		}
	}
	return []any{result}
}

func isNotFoundError(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}
