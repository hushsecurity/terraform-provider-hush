package mcp_application

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	assignAllDesc  = "Grant the application to every user of the organization, whatever `assignment` says."
	assignmentDesc = "A rule granting the application to the users and agents it matches. An " +
		"agent may use the application when any of the rules matches it; a rule matches when " +
		"all of its `condition` blocks do. With neither this nor `assign_all`, the application " +
		"is granted to no one, and users can only request access to it."
	conditionDesc = "A condition of the rule. It holds when any of its `match` blocks does."
	matchDesc     = "One comparison of a property of the user or of the agent."
	sourceDesc    = "What the property belongs to: 'user' (a claim of the user's identity " +
		"provider profile, or 'groups' for the groups the user is in) or 'agent'."
	propertyDesc = "The property compared: for 'user', a claim such as 'email' or 'groups'; " +
		"for 'agent', only 'id'."
	opDesc = "How the property is compared: 'eq' and 'neq' against `value`, 'in' and 'nin' " +
		"against `values`."
	valueDesc  = "The value 'eq' and 'neq' compare against."
	valuesDesc = "The values 'in' and 'nin' compare against. At least one."
)

func assignmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"assign_all": {
			Description: assignAllDesc,
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"assignment": {
			Description: assignmentDesc,
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"condition": {
					Description: conditionDesc,
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"match": {
							Description: matchDesc,
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"source": {
									Description:  sourceDesc,
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice(client.AssignmentSources, false),
								},
								"property": {
									Description:  propertyDesc,
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringIsNotEmpty,
								},
								"op": {
									Description:  opDesc,
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice(client.AssignmentOps, false),
								},
								"value": {
									Description: valueDesc,
									Type:        schema.TypeString,
									Optional:    true,
								},
								"values": {
									Description: valuesDesc,
									Type:        schema.TypeList,
									Optional:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
							}},
						},
					}},
				},
			}},
		},
	}
}

func listOp(op string) bool { return op == "in" || op == "nin" }

// validateAssignments holds each match to what its op takes, and an agent
// match to the one agent property heimdall knows. The schema cannot say
// either: both depend on another member of the same block.
func validateAssignments(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("assignment") {
		return nil
	}
	for i, a := range d.Get("assignment").([]any) {
		if a == nil {
			continue
		}
		for j, c := range a.(map[string]any)["condition"].([]any) {
			if c == nil {
				continue
			}
			for k, m := range c.(map[string]any)["match"].([]any) {
				if m == nil {
					continue
				}
				prefix := fmt.Sprintf("assignment[%d].condition[%d].match[%d]", i, j, k)
				if err := validateMatch(d, prefix, fmt.Sprintf("assignment.%d.condition.%d.match.%d", i, j, k), m.(map[string]any)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateMatch(d *schema.ResourceDiff, prefix, key string, match map[string]any) error {
	if d.NewValueKnown(key+".source") && d.NewValueKnown(key+".property") &&
		match["source"] == "agent" && match["property"] != "id" {
		return fmt.Errorf("%s: an agent match can only compare property %q", prefix, "id")
	}
	if !d.NewValueKnown(key+".op") || !d.NewValueKnown(key+".value") || !d.NewValueKnown(key+".values") {
		return nil
	}
	op := match["op"].(string)
	value, _ := match["value"].(string)
	values, _ := match["values"].([]any)
	if listOp(op) {
		if value != "" {
			return fmt.Errorf("%s: op %q compares against values, not value", prefix, op)
		}
		if len(values) == 0 {
			return fmt.Errorf("%s: op %q needs at least one entry in values", prefix, op)
		}
		return nil
	}
	if len(values) > 0 {
		return fmt.Errorf("%s: op %q compares against value, not values", prefix, op)
	}
	if value == "" {
		return fmt.Errorf("%s: op %q needs value", prefix, op)
	}
	return nil
}

func expandAssignments(d *schema.ResourceData) []client.Assignment {
	out := make([]client.Assignment, 0)
	for _, a := range d.Get("assignment").([]any) {
		if a == nil {
			continue
		}
		var assignment client.Assignment
		for _, c := range a.(map[string]any)["condition"].([]any) {
			if c == nil {
				continue
			}
			var condition client.AssignmentCondition
			for _, m := range c.(map[string]any)["match"].([]any) {
				if m == nil {
					continue
				}
				match := m.(map[string]any)
				predicate := client.AssignmentPredicate{
					Source:   match["source"].(string),
					Property: match["property"].(string),
					Op:       match["op"].(string),
					Value:    match["value"].(string),
				}
				if listOp(predicate.Op) {
					values := make([]string, 0)
					for _, v := range match["values"].([]any) {
						if s, ok := v.(string); ok {
							values = append(values, s)
						}
					}
					predicate.Value = values
				}
				condition.AnyOf = append(condition.AnyOf, predicate)
			}
			assignment.AllOf = append(assignment.AllOf, condition)
		}
		out = append(out, assignment)
	}
	return out
}

func flattenAssignments(assignments []client.Assignment) []any {
	out := make([]any, 0, len(assignments))
	for _, assignment := range assignments {
		conditions := make([]any, 0, len(assignment.AllOf))
		for _, condition := range assignment.AllOf {
			matches := make([]any, 0, len(condition.AnyOf))
			for _, predicate := range condition.AnyOf {
				match := map[string]any{
					"source":   predicate.Source,
					"property": predicate.Property,
					"op":       predicate.Op,
					"value":    "",
					"values":   []any{},
				}
				switch v := predicate.Value.(type) {
				case string:
					match["value"] = v
				case []any:
					match["values"] = v
				}
				matches = append(matches, match)
			}
			conditions = append(conditions, map[string]any{"match": matches})
		}
		out = append(out, map[string]any{"condition": conditions})
	}
	return out
}
