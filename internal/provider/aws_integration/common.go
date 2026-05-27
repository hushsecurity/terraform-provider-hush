package aws_integration

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc                    = "The unique identifier of the AWS integration"
	nameDesc                  = "The name of the AWS integration"
	descriptionDesc           = "The description of the AWS integration"
	roleArnDesc               = "The IAM role ARN for Hush to assume. Mutually exclusive with `cf_stackset_arn`."
	cfStacksetArnDesc         = "The CloudFormation StackSet ARN. Mutually exclusive with `role_arn`."
	uniqueSuffixDesc          = "A unique suffix for CloudFormation resources. Required when using `cf_stackset_arn`."
	accountIDsDesc            = "The list of AWS account IDs discovered via this integration"
	cfStackIDDesc             = "The CloudFormation stack ID (if applicable)"
	statusDesc                = "The current status of the integration"
	featuresDesc              = "List of AWS features and their states"
	featureNameDesc           = "The feature name (secrets_manager, ssm_parameter_store, s3_tf_state, ecr, iam)"
	featureStateDesc          = "The current state of the feature (enabled, disabled, warning, error)"
	featureStateMessageDesc   = "Additional details about the feature state"
	featureAllowedRegionsDesc = "The AWS regions allowed for this feature"
)

func AWSIntegrationResourceSchema() map[string]*schema.Schema {
	s := AWSIntegrationDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 60),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 200),
	}
	s["role_arn"] = &schema.Schema{
		Description:  roleArnDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws:iam::\d{12}:role/`), "must be a valid IAM role ARN"),
		ExactlyOneOf: []string{"role_arn", "cf_stackset_arn"},
	}
	s["cf_stackset_arn"] = &schema.Schema{
		Description:  cfStacksetArnDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"role_arn", "cf_stackset_arn"},
	}
	s["unique_suffix"] = &schema.Schema{
		Description:  uniqueSuffixDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"cf_stackset_arn"},
	}

	return s
}

func AWSIntegrationDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description:   idDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"id"},
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role_arn": {
			Description: roleArnDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cf_stackset_arn": {
			Description: cfStacksetArnDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cf_stack_id": {
			Description: cfStackIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"unique_suffix": {
			Description: uniqueSuffixDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"account_ids": {
			Description: accountIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"features": {
			Description: featuresDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: featureNameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state": {
						Description: featureStateDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state_message": {
						Description: featureStateMessageDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"allowed_regions": {
						Description: featureAllowedRegionsDesc,
						Type:        schema.TypeList,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
	}
}

func awsIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.AWSIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetAWSIntegration(ctx, c, id)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	} else if id, exists := d.GetOk("id"); exists {
		integrationID := id.(string)
		integration, err = client.GetAWSIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no AWS integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetAWSIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup AWS integration by name '%s': %w", integrationName, lookupErr))
		}
		if len(integrations) == 0 {
			return diag.Errorf("no AWS integration found with name: %s", integrationName)
		}
		if len(integrations) > 1 {
			return diag.Errorf("multiple AWS integrations found with name: %s, please use id instead", integrationName)
		}
		// Get full details
		integration, err = client.GetAWSIntegration(ctx, c, integrations[0].ID)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		return diag.Errorf("one of `id` or `name` must be specified")
	}

	d.SetId(integration.ID)

	fields := map[string]any{
		"name":            integration.Name,
		"description":     integration.Description,
		"role_arn":        integration.RoleArn,
		"cf_stackset_arn": integration.CfStacksetArn,
		"cf_stack_id":     integration.CfStackID,
		"unique_suffix":   integration.UniqueSuffix,
		"account_ids":     integration.AccountIDs,
		"status":          integration.Status,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if integration.Features != nil {
		features := make([]map[string]any, len(integration.Features))
		for i, f := range integration.Features {
			features[i] = map[string]any{
				"name":            f.Name,
				"state":           f.State,
				"state_message":   f.StateMessage,
				"allowed_regions": f.AllowedRegions,
			}
		}
		if err := d.Set("features", features); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set features: %w", err))
		}
	}

	return nil
}
