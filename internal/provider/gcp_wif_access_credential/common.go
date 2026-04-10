package gcp_wif_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                 = "The unique identifier of the GCP WIF access credential"
	nameDesc               = "The name of the GCP WIF access credential"
	descriptionDesc        = "The description of the GCP WIF access credential"
	deploymentIDsDesc      = "List of deployment IDs that can access this credential"
	projectNumberDesc      = "The GCP project number"
	poolIDDesc             = "The workload identity pool ID"
	workloadProviderIDDesc = "The workload identity provider ID"
	audienceDesc           = "The audience for the GCP WIF access credential"
	issuerURLDesc          = "The issuer URL for the GCP WIF access credential"
	typeDesc               = "The type of access credential"
	kindDesc               = "The kind of access credential"
)

func ResourceSchema() map[string]*schema.Schema {
	s := DataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 1000),
	}
	s["deployment_ids"] = &schema.Schema{
		Description: deploymentIDsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["project_number"] = &schema.Schema{
		Description: projectNumberDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["pool_id"] = &schema.Schema{
		Description: poolIDDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["workload_provider_id"] = &schema.Schema{
		Description: workloadProviderIDDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["audience"] = &schema.Schema{
		Description: audienceDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	}

	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: idDesc,
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: nameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"deployment_ids": {
			Description: deploymentIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"project_number": {
			Description: projectNumberDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"pool_id": {
			Description: poolIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"workload_provider_id": {
			Description: workloadProviderIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"audience": {
			Description: audienceDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"issuer_url": {
			Description: issuerURLDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
