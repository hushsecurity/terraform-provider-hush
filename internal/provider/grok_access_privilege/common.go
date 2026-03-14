package grok_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc          = "The unique identifier of the Grok access privilege"
	nameDesc        = "The name of the Grok access privilege"
	descriptionDesc = "The description of the Grok access privilege"
	endpointsDesc   = "The list of allowed Grok API endpoints (e.g., Chat, Batch, Embed). If omitted, all endpoints are allowed."
	modelsDesc      = "The list of allowed Grok model names. If omitted, all models are allowed."
	typeDesc        = "The type of access privilege"
)

var validEndpoints = []string{
	"Chat",
	"Batch",
	"Embed",
	"Files",
	"Image",
	"Models",
	"Sample",
	"Video",
	"Voice",
	"Tokenize",
	"Documents",
}

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
	s["endpoints"] = &schema.Schema{
		Description: endpointsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(validEndpoints, false),
		},
	}
	s["models"] = &schema.Schema{
		Description: modelsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringIsNotEmpty,
		},
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
		"endpoints": {
			Description: endpointsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"models": {
			Description: modelsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
