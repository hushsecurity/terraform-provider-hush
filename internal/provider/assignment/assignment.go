// Package assignment is the rule blocks that grant something to users (and
// agents): an application, or a service agent's users.
package assignment

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	conditionDesc = "A condition of the rule. It holds when any of its `match` blocks does."
	opDesc        = "How the property is compared: 'eq' and 'neq' against `value`, 'in' and 'nin' " +
		"against `values`."
	valueDesc  = "The value 'eq' and 'neq' compare against."
	valuesDesc = "The values 'in' and 'nin' compare against. At least one."
)

// Schema is a list of rules: a rule matches when all of its `condition`
// blocks do. sources are what a match may look at.
func Schema(desc, matchDesc, sourceDesc, propertyDesc string, sources []string) *schema.Schema {
	return &schema.Schema{
		Description: desc,
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
								ValidateFunc: validation.StringInSlice(sources, false),
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
	}
}

func listOp(op string) bool { return op == "in" || op == "nin" }

// Validate holds each match of the `key` rules to what its op takes, and an
// agent match to the one agent property heimdall knows. The schema cannot say
// either: both depend on another member of the same block.
func Validate(d *schema.ResourceDiff, key string) error {
	if !d.NewValueKnown(key) {
		return nil
	}
	for i, a := range d.Get(key).([]any) {
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
				prefix := fmt.Sprintf("%s[%d].condition[%d].match[%d]", key, i, j, k)
				if err := validateMatch(d, prefix, fmt.Sprintf("%s.%d.condition.%d.match.%d", key, i, j, k), m.(map[string]any)); err != nil {
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

// Expand reads the `key` rules.
func Expand(d interface{ Get(string) any }, key string) []client.Assignment {
	out := make([]client.Assignment, 0)
	for _, a := range d.Get(key).([]any) {
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

// Flatten is the rules as their blocks.
func Flatten(assignments []client.Assignment) []any {
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
