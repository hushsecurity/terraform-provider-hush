package elasticsearch_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the Elasticsearch access privilege"
	nameDesc        = "The name of the Elasticsearch access privilege"
	descriptionDesc = "The description of the Elasticsearch access privilege"
	grantDesc       = "The Elasticsearch grant configuration"
	clusterDesc     = "The list of cluster-level privileges (e.g., monitor, manage, all)"
	indicesDesc     = "The list of index-level privilege definitions"
	namesDesc       = "The list of index name patterns (e.g., \"*\", \"logs-*\")"
	privilegesDesc  = "The list of index-level privileges (e.g., read, write, all)"
	typeDesc        = "The type of access privilege"
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
	s["grant"] = &schema.Schema{
		Description: grantDesc,
		Type:        schema.TypeList,
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster": {
					Description: clusterDesc,
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"indices": {
					Description: indicesDesc,
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"names": {
								Description: namesDesc,
								Type:        schema.TypeList,
								Required:    true,
								MinItems:    1,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"privileges": {
								Description: privilegesDesc,
								Type:        schema.TypeList,
								Required:    true,
								MinItems:    1,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
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
		"grant": {
			Description: grantDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cluster": {
						Description: clusterDesc,
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"indices": {
						Description: indicesDesc,
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"names": {
									Description: namesDesc,
									Type:        schema.TypeList,
									Computed:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"privileges": {
									Description: privilegesDesc,
									Type:        schema.TypeList,
									Computed:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func expandGrant(list []any) client.ElasticsearchGrant {
	if len(list) == 0 {
		return client.ElasticsearchGrant{}
	}
	m := list[0].(map[string]any)
	grant := client.ElasticsearchGrant{}

	if v, ok := m["cluster"]; ok {
		clusterList := v.([]any)
		cluster := make([]string, len(clusterList))
		for i, item := range clusterList {
			cluster[i] = item.(string)
		}
		grant.Cluster = cluster
	}

	if v, ok := m["indices"]; ok {
		indicesList := v.([]any)
		indices := make([]client.ElasticsearchIndexPrivilege, len(indicesList))
		for i, item := range indicesList {
			idx := item.(map[string]any)
			namesList := idx["names"].([]any)
			names := make([]string, len(namesList))
			for j, n := range namesList {
				names[j] = n.(string)
			}
			privsList := idx["privileges"].([]any)
			privs := make([]string, len(privsList))
			for j, p := range privsList {
				privs[j] = p.(string)
			}
			indices[i] = client.ElasticsearchIndexPrivilege{
				Names:      names,
				Privileges: privs,
			}
		}
		grant.Indices = indices
	}

	return grant
}

func flattenGrant(grant client.ElasticsearchGrant) []any {
	m := map[string]any{}

	if grant.Cluster != nil {
		m["cluster"] = grant.Cluster
	} else {
		m["cluster"] = []string{}
	}

	if grant.Indices != nil {
		indices := make([]any, len(grant.Indices))
		for i, idx := range grant.Indices {
			indices[i] = map[string]any{
				"names":      idx.Names,
				"privileges": idx.Privileges,
			}
		}
		m["indices"] = indices
	} else {
		m["indices"] = []any{}
	}

	return []any{m}
}
