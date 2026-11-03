package pinot_access_credential

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the Pinot access credential"
	nameDesc          = "The name of the Pinot access credential"
	descriptionDesc   = "The description of the Pinot access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	hostDesc          = "The hostname or IP address of the Pinot controller. Hush writes users through its `/users` API, so the cluster must keep its users in ZooKeeper (`ZkBasicAuthAccessControlFactory`)"
	portDesc          = "The port number of the Pinot controller (default: 9000)"
	usernameDesc      = "The username of an `ADMIN` user on the Pinot controller. It must be able to read the controller's `/appconfigs`, which is how Hush checks the access control"
	passwordDesc      = "The password of the Pinot controller user"
	passwordWODesc    = "The password of the Pinot controller user (write-only). This is a write-only attribute that is more secure than `password` because Terraform will not store this value in the state file. Either `password` or `password_wo` must be specified."
	passwordWOVerDesc = "Used to trigger updates for `password_wo`. This value should be changed when the password content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	tlsDesc           = "Whether to use TLS for the Pinot controller. `broker_url` carries its own scheme. Set without `tls_ca`, it verifies neither the controller's certificate nor that of an `https` `broker_url`"
	tlsCADesc         = "The TLS CA certificate that verifies the Pinot controller and an `https` `broker_url`, in place of the system CAs, so both need certificates from it"
	brokerURLDesc     = "The `http(s)` URL of a Pinot broker (e.g., `http://pinot-broker:8099`), with no credentials, query, fragment or whitespace. If set, Hush checks that the broker accepts each generated user, and delivery templates can use it as `${broker_url}`. Behind a load balancer the check reaches one broker only. A trailing `/` is dropped"
	typeDesc          = "The type of access credential"
	kindDesc          = "The kind of access credential"
	secretStoreIDDesc = "The ID of the secret store where this credential is saved (optional)"
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
		Description: deploymentIDsDesc + ". Changing this after creation is not supported; the credential must be deleted and recreated.",
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		MaxItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["secret_store_id"] = &schema.Schema{
		Description:  secretStoreIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^sst-`), "secret_store_id must start with 'sst-'"),
	}
	s["host"] = &schema.Schema{
		Description: hostDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["port"] = &schema.Schema{
		Description:  portDesc,
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      9000,
		ValidateFunc: validation.IsPortNumber,
	}
	s["username"] = &schema.Schema{
		Description:  usernameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
	}
	s["password"] = &schema.Schema{
		Description:   passwordDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ValidateFunc:  validation.StringLenBetween(1, 256),
		ConflictsWith: []string{"password_wo"},
		ExactlyOneOf:  []string{"password", "password_wo"},
	}
	s["password_wo"] = &schema.Schema{
		Description:   passwordWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ValidateFunc:  validation.StringLenBetween(1, 256),
		ConflictsWith: []string{"password"},
		ExactlyOneOf:  []string{"password", "password_wo"},
		RequiredWith:  []string{"password_wo_version"},
	}
	s["password_wo_version"] = &schema.Schema{
		Description:  passwordWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"password_wo"},
	}
	s["tls"] = &schema.Schema{
		Description: tlsDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	}
	s["tls_ca"] = &schema.Schema{
		Description: tlsCADesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["broker_url"] = &schema.Schema{
		Description:      brokerURLDesc,
		Type:             schema.TypeString,
		Optional:         true,
		ValidateFunc:     validateBrokerURL,
		DiffSuppressFunc: suppressTrailingSlash,
	}

	return s
}

// validateBrokerURL mirrors midgard, which refuses whitespace and control
// characters, then strips trailing slashes before it checks the URL.
// ${broker_url}/query/sql cannot be built from a query or a fragment, and
// userinfo would put a password in the plan and the state.
func validateBrokerURL(v any, k string) ([]string, []error) {
	if strings.IndexFunc(v.(string), func(r rune) bool {
		return unicode.IsSpace(r) || !unicode.IsPrint(r)
	}) >= 0 {
		return nil, []error{fmt.Errorf("%s cannot contain whitespace or control characters", k)}
	}
	s := strings.TrimRight(v.(string), "/")
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, []error{fmt.Errorf("%s must be an http(s) URL, got %q", k, v)}
	}
	if u.User != nil {
		return nil, []error{fmt.Errorf("%s cannot carry a username or password", k)}
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.HasSuffix(s, "#") {
		return nil, []error{fmt.Errorf("%s cannot carry a query or a fragment", k)}
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err != nil || n > 65535 {
			return nil, []error{fmt.Errorf("%s has an invalid port %q", k, p)}
		}
	}
	return nil, nil
}

// suppressTrailingSlash: the API stores broker_url without trailing slashes, so
// a configured one would otherwise differ from state on every plan.
func suppressTrailingSlash(_, old, new string, _ *schema.ResourceData) bool {
	return old != "" && strings.TrimRight(new, "/") == old
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
		"host": {
			Description: hostDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: portDesc,
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"username": {
			Description: usernameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tls": {
			Description: tlsDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"tls_ca": {
			Description: tlsCADesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"broker_url": {
			Description: brokerURLDesc,
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
		"secret_store_id": {
			Description: secretStoreIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
