package mcp_application

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	toolDefaultsDesc = "What the gateway does with a call to a tool of each class, unless " +
		"`tool_operation` sets the tool's own. Each is 'allow', 'block' or 'user_consent' (ask " +
		"the user, through the deployment's `hush_agw_consent_methods`). A class left unset keeps " +
		"what the application has, which starts as allow for read, user_consent for write and " +
		"block for destructive; removing the block changes nothing."
	toolDefaultReadDesc        = "The operation for tools of class 'read'."
	toolDefaultWriteDesc       = "The operation for tools of class 'write'."
	toolDefaultDestructiveDesc = "The operation for tools of class 'destructive'."
	toolOperationBlockDesc     = "The operation for one tool, overriding its class default. The " +
		"tool must be one of the application's (see `tools`). A tool without a block takes its " +
		"class default, so removing a block returns the tool to it, and an override set " +
		"elsewhere, such as in the console, is removed on the next apply. On create the name is " +
		"checked at plan time against the tools of the entry or custom app as they stand, so " +
		"a tool added to a custom app in the same apply is refused: apply the custom app first."
	toolOperationNameDesc = "The tool's name."
	toolOperationOpDesc   = "'allow', 'block' or 'user_consent'."
)

func toolOperationsSchema() map[string]*schema.Schema {
	operation := func(desc string) *schema.Schema {
		return &schema.Schema{
			Description:  desc,
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(client.ToolOperations, false),
		}
	}
	return map[string]*schema.Schema{
		"tool_defaults": {
			Description: toolDefaultsDesc,
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"read":        operation(toolDefaultReadDesc),
				"write":       operation(toolDefaultWriteDesc),
				"destructive": operation(toolDefaultDestructiveDesc),
			}},
		},
		"tool_operation": {
			Description: toolOperationBlockDesc,
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Description:  toolOperationNameDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				"operation": {
					Description:  toolOperationOpDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(client.ToolOperations, false),
				},
			}},
		},
	}
}

// planToolsRecompute marks tools unknown when an edit changes what they
// resolve to. Their operation follows tool_defaults and tool_operation, and
// the plan would otherwise show the old ones as settled values to anything
// that reads them.
func planToolsRecompute(d *schema.ResourceDiff) error {
	if d.Id() == "" || !d.HasChanges("tool_defaults", "tool_operation") {
		return nil
	}
	return d.SetNewComputed("tools")
}

// validateToolOperations refuses two blocks for one tool: a set tells blocks
// apart by their whole content, so both would stand and only one could win.
func validateToolOperations(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("tool_operation") {
		return nil
	}
	seen := map[string]bool{}
	for _, v := range d.Get("tool_operation").(*schema.Set).List() {
		name := v.(map[string]any)["name"].(string)
		if name == "" {
			continue
		}
		if seen[name] {
			return fmt.Errorf("tool_operation: tool %q has more than one block", name)
		}
		seen[name] = true
	}
	return nil
}

func toolOperationMap(set any) map[string]string {
	out := map[string]string{}
	if set == nil {
		return out
	}
	for _, v := range set.(*schema.Set).List() {
		block := v.(map[string]any)
		out[block["name"].(string)] = block["operation"].(string)
	}
	return out
}

// applyToolOperations sends what changed. The group defaults go first: a tool
// with no setting of its own follows them. A tool whose block was removed is
// sent a null operation, which returns it to its class default.
func applyToolOperations(ctx context.Context, c *client.Client, d *schema.ResourceData, creating bool) error {
	if creating || d.HasChange("tool_defaults") {
		var groups []client.ToolGroupOperationChange
		for _, toolType := range client.ToolTypes {
			key := "tool_defaults.0." + toolType
			if !creating && !d.HasChange(key) {
				continue
			}
			if operation, _ := d.Get(key).(string); operation != "" {
				groups = append(groups, client.ToolGroupOperationChange{Type: toolType, Operation: operation})
			}
		}
		if len(groups) > 0 {
			if _, err := client.ChangeMCPToolGroupOperations(ctx, c, d.Id(), groups); err != nil {
				return fmt.Errorf("tool_defaults: %w", err)
			}
		}
	}

	if !creating && !d.HasChange("tool_operation") {
		return nil
	}
	previous, current := d.GetChange("tool_operation")
	before, after := toolOperationMap(previous), toolOperationMap(current)
	var changes []client.ToolOperationChange
	for name, operation := range after {
		if before[name] != operation {
			changes = append(changes, client.ToolOperationChange{Name: name, Operation: &operation})
		}
	}
	for name := range before {
		if _, kept := after[name]; !kept {
			changes = append(changes, client.ToolOperationChange{Name: name})
		}
	}
	if len(changes) == 0 {
		return nil
	}
	// Stable order, so one request reads the same from one run to the next.
	slices.SortFunc(changes, func(a, b client.ToolOperationChange) int {
		return cmp.Compare(a.Name, b.Name)
	})
	if _, err := client.ChangeMCPToolOperations(ctx, c, d.Id(), changes); err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return fmt.Errorf("tool_operation: the application has no such tool (see its tools attribute): %w", err)
		}
		return fmt.Errorf("tool_operation: %w", err)
	}
	return nil
}

func flattenToolDefaults(groups []client.ToolGroup) []any {
	block := map[string]any{}
	for _, group := range groups {
		block[group.Type] = group.Operation
	}
	return []any{block}
}

func flattenToolOperations(tools []client.MCPTool) []any {
	out := make([]any, 0)
	for _, tool := range tools {
		if tool.Operation != nil {
			out = append(out, map[string]any{"name": tool.Name, "operation": *tool.Operation})
		}
	}
	return out
}
