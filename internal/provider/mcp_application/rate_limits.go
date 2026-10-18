package mcp_application

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	rateLimitDesc = "A cap on how many calls each caller (a user, or a service agent) may make " +
		"through the gateway: to any of the application's tools, to its tools of one class, or " +
		"to one tool. A call must fit every limit it matches. One block per target; a limit " +
		"without a block is removed, including one set elsewhere, such as in the console."
	rateLimitTargetTypeDesc = "'app', 'group' or 'tool'."
	rateLimitTargetDesc     = "For 'group', the class: 'read', 'write' or 'destructive'. For " +
		"'tool', the tool's name (see `tools`). Unset for 'app'."
	rateLimitMaxCallsDesc = "The calls allowed at once; they come back one by one over `window`."
	rateLimitWindowDesc   = "Seconds, from 60 to 86400."
	rateLimitModeDesc     = "'block' refuses calls past the limit; 'warn' lets them through and " +
		"flags them in the activity log."
)

func rateLimitSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"rate_limit": {
			Description: rateLimitDesc,
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"target_type": {
					Description:  rateLimitTargetTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(client.RateLimitTargetTypes, false),
				},
				"target": {
					Description: rateLimitTargetDesc,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"max_calls": {
					Description:  rateLimitMaxCallsDesc,
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
				"window": {
					Description:  rateLimitWindowDesc,
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(60, 86400),
				},
				"mode": {
					Description:  rateLimitModeDesc,
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "block",
					ValidateFunc: validation.StringInSlice(client.RateLimitModes, false),
				},
			}},
		},
	}
}

type rateLimitKey struct{ targetType, target string }

func rateLimitMap(set any) map[rateLimitKey]client.RateLimitIn {
	out := map[rateLimitKey]client.RateLimitIn{}
	if set == nil {
		return out
	}
	for _, v := range set.(*schema.Set).List() {
		block := v.(map[string]any)
		key := rateLimitKey{block["target_type"].(string), block["target"].(string)}
		out[key] = client.RateLimitIn{
			MaxCalls: block["max_calls"].(int),
			Window:   block["window"].(int),
			Mode:     block["mode"].(string),
		}
	}
	return out
}

// validateRateLimits refuses what heimdall would refuse at apply: a target
// with two blocks, or a target that does not fit its type. A target another
// resource supplies reads as empty until known, so empty is left to apply.
func validateRateLimits(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("rate_limit") {
		return nil
	}
	seen := map[rateLimitKey]bool{}
	for _, v := range d.Get("rate_limit").(*schema.Set).List() {
		block := v.(map[string]any)
		key := rateLimitKey{block["target_type"].(string), block["target"].(string)}
		switch key.targetType {
		case "app":
			if key.target != "" {
				return errors.New("rate_limit: an 'app' limit takes no target")
			}
		case "group":
			if key.target != "" && !slices.Contains(client.ToolTypes, key.target) {
				return fmt.Errorf("rate_limit: group %q is not one of %v", key.target, client.ToolTypes)
			}
		}
		if key.target == "" && key.targetType != "app" {
			continue
		}
		if seen[key] {
			return fmt.Errorf("rate_limit: %s %q has more than one block", key.targetType, key.target)
		}
		seen[key] = true
	}
	return nil
}

// applyRateLimits sets every limit that is new or changed, and removes those
// whose block is gone.
func applyRateLimits(ctx context.Context, c *client.Client, d *schema.ResourceData) error {
	if !d.HasChange("rate_limit") {
		return nil
	}
	previous, current := d.GetChange("rate_limit")
	before, after := rateLimitMap(previous), rateLimitMap(current)
	for key, limit := range after {
		if old, ok := before[key]; ok && old == limit {
			continue
		}
		if err := client.SetMCPRateLimit(ctx, c, d.Id(), key.targetType, key.target, &limit); err != nil {
			var apiErr *client.APIError
			if key.targetType == "tool" && errors.As(err, &apiErr) && apiErr.IsNotFound() {
				return fmt.Errorf("rate_limit: the application has no tool %q (see its tools attribute): %w", key.target, err)
			}
			return fmt.Errorf("rate_limit: %w", err)
		}
	}
	for key := range before {
		if _, kept := after[key]; kept {
			continue
		}
		if err := client.SetMCPRateLimit(ctx, c, d.Id(), key.targetType, key.target, nil); err != nil {
			return fmt.Errorf("rate_limit: %w", err)
		}
	}
	return nil
}

func flattenRateLimits(limits []client.RateLimit) []any {
	out := make([]any, 0, len(limits))
	for _, rl := range limits {
		target := ""
		if rl.Target != nil {
			target = *rl.Target
		}
		out = append(out, map[string]any{
			"target_type": rl.TargetType,
			"target":      target,
			"max_calls":   rl.MaxCalls,
			"window":      rl.Window,
			"mode":        rl.Mode,
		})
	}
	return out
}
