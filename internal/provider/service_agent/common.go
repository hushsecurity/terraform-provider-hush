package service_agent

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	resourceDescription = "A service agent: an AI agent that signs in to Hush with the identity of the " +
		"workload it runs on (an AWS IAM role, a Kubernetes service account, an Azure / Entra service " +
		"principal or any OIDC issuer) instead of a person's sign-in. It acts as itself or, when " +
		"`acts_for_users` is on, for a signed-in user. Give it apps with a `hush_mcp_application` " +
		"assignment rule on `agent.id`: assigning an app to all users does not cover service agents. " +
		"A client secret is managed with `hush_service_agent_client_secret`. An identity can belong " +
		"to one agent at a time: to move it to another agent, remove it from the first in one apply " +
		"and add it to the second in the next."
	nameDesc         = "The agent's name."
	enabledDesc      = "Whether the agent can sign in."
	actsForUsersDesc = "Whether the agent can act for a signed-in user, by exchanging the user's token " +
		"(RFC 8693 token exchange)."
	resourceAttrDesc = "What a front end passes as `resource` when it signs a user in for this agent " +
		"to act for them."
	awsDesc = "An AWS IAM role, signing in with `sts:GetWebIdentityToken` tokens whose audience is " +
		"the deployment's issuer."
	awsIssuerDesc = "The AWS account's outbound identity federation issuer, from " +
		"`aws iam get-outbound-web-identity-federation-info`."
	awsRoleDesc = "The role's ARN. The role needs `sts:GetWebIdentityToken`."
	k8sDesc     = "A Kubernetes service account, signing in with a projected token whose audience is " +
		"the deployment's issuer."
	k8sIssuerDesc   = "The cluster's OIDC issuer, as its `/.well-known/openid-configuration` reports it."
	k8sNSDesc       = "The service account's namespace."
	k8sSADesc       = "The service account's name."
	azureDesc       = "An Azure / Entra service principal: a managed identity, an app registration or an Entra agent identity (e.g. a Foundry agent). It signs in with an app-only Entra token for `audience`, so it also sends `resource` (the gateway) when it signs in. The app registration must issue v2 tokens and be dedicated to Hush: anything that receives tokens for it could replay them."
	azureTenantDesc = "The Entra tenant id, lowercase."
	azureSPDesc     = "The service principal's object id (the token's `sub`), lowercase."
	azureAudDesc    = "The application (client) id of the app registration the identity requests tokens for, lowercase."
	oidcDesc        = "Any OIDC issuer: a token from `issuer` whose `sub` is `subject`, with the deployment's issuer as its audience."
	oidcIssuerDesc  = "The issuer, an https URL whose `/.well-known/openid-configuration` is public."
	oidcSubjectDesc = "The token's `sub`."
	conditionsDesc  = "Extra checks on the token's claims; all must hold."
	claimDesc       = "The claim: its name, or a path into nested claims (up to 5 names)."
	opDesc          = "'eq' (equal), 'pfx' (starts with) or 'sfx' (ends with)."
	valueDesc       = "The value to compare with. A list claim matches if any element does."
)

var (
	awsRolePattern = regexp.MustCompile(`^arn:aws[a-z-]*:iam::\d{12}:role/[\w+=,.@/-]+$`)
	k8sNSPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	k8sSAPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`)
	guidPattern    = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// credentialFields are each federated type's own fields, in the order they
// are written in a block.
var credentialFields = map[string][]string{
	client.CredentialAWSIAMRole: {"issuer", "role_arn"},
	client.CredentialKubernetes: {"issuer", "namespace", "service_account"},
	client.CredentialAzure:      {"tenant_id", "service_principal_id", "audience"},
	client.CredentialOIDC:       {"issuer", "subject"},
}

func conditionsSchema() *schema.Schema {
	return &schema.Schema{
		Description: conditionsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    10,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"claim": {
				Description: claimDesc,
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    5,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"op": {
				Description:  opDesc,
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "eq",
				ValidateFunc: validation.StringInSlice(client.ClaimOps, false),
			},
			"value": {
				Description:  valueDesc,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		}},
	}
}

func str(desc string, validate func(any, string) ([]string, []error)) *schema.Schema {
	return &schema.Schema{Description: desc, Type: schema.TypeString, Required: true, ValidateFunc: validate}
}

func credentialBlock(desc string, fields map[string]*schema.Schema) *schema.Schema {
	fields["conditions"] = conditionsSchema()
	return &schema.Schema{
		Description: desc,
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Resource{Schema: fields},
	}
}

func ResourceSchema() map[string]*schema.Schema {
	https := validation.IsURLWithHTTPS
	guid := validation.StringMatch(guidPattern, "must be a lowercase GUID")
	return map[string]*schema.Schema{
		"name": {
			Description:  nameDesc,
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 128),
		},
		"enabled":        {Description: enabledDesc, Type: schema.TypeBool, Optional: true, Default: true},
		"acts_for_users": {Description: actsForUsersDesc, Type: schema.TypeBool, Optional: true, Default: false},
		"resource":       {Description: resourceAttrDesc, Type: schema.TypeString, Computed: true},
		client.CredentialAWSIAMRole: credentialBlock(awsDesc, map[string]*schema.Schema{
			"issuer":   str(awsIssuerDesc, https),
			"role_arn": str(awsRoleDesc, validation.StringMatch(awsRolePattern, "must be an IAM role ARN")),
		}),
		client.CredentialKubernetes: credentialBlock(k8sDesc, map[string]*schema.Schema{
			"issuer":          str(k8sIssuerDesc, https),
			"namespace":       str(k8sNSDesc, validation.StringMatch(k8sNSPattern, "must be a Kubernetes namespace name")),
			"service_account": str(k8sSADesc, validation.StringMatch(k8sSAPattern, "must be a Kubernetes service account name")),
		}),
		client.CredentialAzure: credentialBlock(azureDesc, map[string]*schema.Schema{
			"tenant_id":            str(azureTenantDesc, guid),
			"service_principal_id": str(azureSPDesc, guid),
			"audience":             str(azureAudDesc, guid),
		}),
		client.CredentialOIDC: credentialBlock(oidcDesc, map[string]*schema.Schema{
			"issuer":  str(oidcIssuerDesc, https),
			"subject": str(oidcSubjectDesc, validation.StringLenBetween(1, 2048)),
		}),
	}
}

// identityFields are the fields that name the identity; the rest of a block
// (its conditions, an Azure audience) can change while the identity stays.
var identityFields = map[string][]string{
	client.CredentialAWSIAMRole: {"issuer", "role_arn"},
	client.CredentialKubernetes: {"issuer", "namespace", "service_account"},
	client.CredentialAzure:      {"tenant_id", "service_principal_id"},
	client.CredentialOIDC:       {"issuer", "subject"},
}

// credential is a federated credential keyed two ways: key by everything in
// it (a changed block is a different credential), identity by the issuer and
// subject heimdall binds once per organization, whatever the block's type.
type credential struct {
	key      string
	identity string
	body     map[string]any
	id       string // heimdall's id, for one it holds
}

func newCredential(body map[string]any) (credential, error) {
	key, err := credentialKey(body)
	if err != nil {
		return credential{}, err
	}
	issuer, subject := identityOf(body)
	return credential{key: key, identity: issuer + " " + subject, body: body}, nil
}

// identityOf derives the issuer and subject as heimdall does, so a typed block
// and the oidc block naming the same identity collide here too.
func identityOf(body map[string]any) (issuer, subject string) {
	field := func(name string) string { v, _ := body[name].(string); return v }
	switch body["type"] {
	case client.CredentialAWSIAMRole:
		return field("issuer"), field("role_arn")
	case client.CredentialKubernetes:
		return field("issuer"), "system:serviceaccount:" + field("namespace") + ":" + field("service_account")
	case client.CredentialAzure:
		return "https://login.microsoftonline.com/" + field("tenant_id") + "/v2.0", field("service_principal_id")
	default:
		return field("issuer"), field("subject")
	}
}

// getter is what both a plan's diff and an applied resource read from.
type getter interface {
	Get(string) any
}

// configuredCredentials reads every federated credential block.
func configuredCredentials(d getter) ([]credential, error) {
	var out []credential
	for _, kind := range client.FederatedCredentialTypes {
		for _, raw := range d.Get(kind).(*schema.Set).List() {
			block, _ := raw.(map[string]any)
			body := map[string]any{"type": kind}
			for _, field := range credentialFields[kind] {
				body[field], _ = block[field].(string)
			}
			conditions, _ := block["conditions"].([]any)
			if conditions := expandConditions(conditions); len(conditions) > 0 {
				body["conditions"] = conditions
			}
			c, err := newCredential(body)
			if err != nil {
				return nil, err
			}
			out = append(out, c)
		}
	}
	return out, nil
}

func expandConditions(blocks []any) []client.ClaimCondition {
	out := make([]client.ClaimCondition, 0, len(blocks))
	for _, raw := range blocks {
		block, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		claims, _ := block["claim"].([]any)
		claim := []string{}
		for _, name := range claims {
			if name, ok := name.(string); ok {
				claim = append(claim, name)
			}
		}
		op, _ := block["op"].(string)
		value, _ := block["value"].(string)
		out = append(out, client.ClaimCondition{Claim: claim, Op: op, Value: value})
	}
	return out
}

// existingCredential keys a credential heimdall holds the same way as a block.
func existingCredential(c client.AgentCredential) (credential, error) {
	body := map[string]any{"type": c.Type}
	values := map[string]string{
		"issuer": c.Issuer, "role_arn": c.RoleARN, "namespace": c.Namespace,
		"service_account": c.ServiceAccount, "tenant_id": c.TenantID,
		"service_principal_id": c.ServicePrincipalID, "audience": c.Audience, "subject": c.Subject,
	}
	for _, field := range credentialFields[c.Type] {
		body[field] = values[field]
	}
	if len(c.Conditions) > 0 {
		body["conditions"] = c.Conditions
	}
	out, err := newCredential(body)
	out.id = c.ID
	return out, err
}

// credentialKey is the body as canonical JSON: map keys sort, and both sides
// carry conditions as client.ClaimCondition.
func credentialKey(body map[string]any) (string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("encoding credential: %w", err)
	}
	return string(data), nil
}

func isFederated(kind string) bool {
	_, ok := credentialFields[kind]
	return ok
}

// flatten writes what heimdall returns. Client secrets have their own
// resource, so they are left out here.
func flatten(d *schema.ResourceData, agent *client.ServiceAgent) error {
	blocks := map[string][]any{}
	for _, kind := range client.FederatedCredentialTypes {
		blocks[kind] = []any{}
	}
	for _, c := range agent.Credentials {
		var block map[string]any
		switch c.Type {
		case client.CredentialAWSIAMRole:
			block = map[string]any{"issuer": c.Issuer, "role_arn": c.RoleARN}
		case client.CredentialKubernetes:
			block = map[string]any{
				"issuer": c.Issuer, "namespace": c.Namespace, "service_account": c.ServiceAccount,
			}
		case client.CredentialAzure:
			block = map[string]any{
				"tenant_id": c.TenantID, "service_principal_id": c.ServicePrincipalID,
				"audience": c.Audience,
			}
		case client.CredentialOIDC:
			block = map[string]any{"issuer": c.Issuer, "subject": c.Subject}
		default:
			continue
		}
		block["conditions"] = flattenConditions(c.Conditions)
		blocks[c.Type] = append(blocks[c.Type], block)
	}

	values := map[string]any{
		"name":           agent.Name,
		"enabled":        agent.Enabled,
		"acts_for_users": agent.ActsForUsers,
		"resource":       agent.Resource,
	}
	for kind, list := range blocks {
		values[kind] = list
	}
	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}
	return nil
}

func flattenConditions(conditions []client.ClaimCondition) []any {
	out := make([]any, 0, len(conditions))
	for _, c := range conditions {
		claim := make([]any, 0, len(c.Claim))
		for _, name := range c.Claim {
			claim = append(claim, name)
		}
		out = append(out, map[string]any{"claim": claim, "op": c.Op, "value": c.Value})
	}
	return out
}

// validateCredentials refuses at plan time what heimdall would refuse halfway
// through an apply: the same identity twice, or more than it allows.
func validateCredentials(d getter) error {
	wanted, err := configuredCredentials(d)
	if err != nil {
		return err
	}
	if len(wanted) > client.MaxAgentCredentials {
		return fmt.Errorf("an agent can have at most %d identities, got %d",
			client.MaxAgentCredentials, len(wanted))
	}
	seen := map[string]bool{}
	for _, c := range wanted {
		if !identityKnown(c) {
			continue // a value from another resource, unknown until apply
		}
		if seen[c.identity] {
			return fmt.Errorf("the same identity is written twice: %s", c.identity)
		}
		seen[c.identity] = true
	}
	return nil
}

func identityKnown(c credential) bool {
	for _, field := range identityFields[c.body["type"].(string)] {
		if c.body[field] == "" {
			return false
		}
	}
	return true
}
