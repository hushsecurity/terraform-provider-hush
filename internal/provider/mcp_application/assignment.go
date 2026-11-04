package mcp_application

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/assignment"
)

const (
	assignAllDesc  = "Grant the application to every user of the organization, whatever `assignment` says."
	assignmentDesc = "A rule granting the application to the users and agents it matches. An " +
		"agent may use the application when any of the rules matches it; a rule matches when " +
		"all of its `condition` blocks do. With neither this nor `assign_all`, the application " +
		"is granted to no one, and users can only request access to it."
	matchDesc  = "One comparison of a property of the user or of the agent."
	sourceDesc = "What the property belongs to: 'user' (a claim of the user's identity " +
		"provider profile, or 'groups' for the groups the user is in) or 'agent'."
	propertyDesc = "The property compared: for 'user', a claim such as 'email' or 'groups'; " +
		"for 'agent', only 'id'."
)

func assignmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"assign_all": {
			Description: assignAllDesc,
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"assignment": assignment.Schema(assignmentDesc, matchDesc, sourceDesc, propertyDesc, client.AssignmentSources),
	}
}
