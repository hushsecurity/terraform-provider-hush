package custom_mcp_application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

// Headers the gateway sets itself, which heimdall refuses in `headers`.
var (
	deniedHeaderNames = map[string]bool{
		"authorization": true, "host": true, "content-length": true, "content-encoding": true,
		"content-type": true, "mcp-session-id": true, "mcp-protocol-version": true,
	}
	deniedHeaderPrefixes = []string{"proxy-", "x-hush-", "hush-"}
)

// customizeDiff refuses at plan time what heimdall refuses at apply and the
// schema cannot say: rules that span list elements, map keys, or members of
// the auth block. A value another resource supplies is unknown here and reads
// as empty, so each rule skips what it cannot see rather than guess.
func customizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if err := validateURLs(d); err != nil {
		return err
	}
	if err := validateHeaders(d); err != nil {
		return err
	}
	return validateAuth(d)
}

// Several addresses are told apart by label, so each needs one, and heimdall
// matches labels without regard to case.
func validateURLs(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("url") {
		return nil
	}
	urls := d.Get("url").([]any)
	if len(urls) < 2 {
		return nil
	}
	seen := map[string]bool{}
	for i := range urls {
		key := fmt.Sprintf("url.%d.label", i)
		if !d.NewValueKnown(key) {
			continue
		}
		label := d.Get(key).(string)
		if label == "" {
			return fmt.Errorf("url[%d]: a label is required when there is more than one url", i)
		}
		if seen[strings.ToLower(label)] {
			return fmt.Errorf("url[%d]: label %q is used more than once", i, label)
		}
		seen[strings.ToLower(label)] = true
	}
	return nil
}

func validateHeaders(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("headers") {
		return nil
	}
	seen := map[string]bool{}
	for name := range d.Get("headers").(map[string]any) {
		lowered := strings.ToLower(name)
		if reason := reservedHeader(name); reason != "" {
			return fmt.Errorf("headers: %s", reason)
		}
		if seen[lowered] {
			return fmt.Errorf("headers: %q is set more than once, ignoring case", name)
		}
		seen[lowered] = true
	}
	return nil
}

func validateAuth(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("auth") || len(d.Get("auth").([]any)) == 0 {
		return nil
	}
	set := func(name string) bool { return writeonly.IsSetNested(d, "auth", 0, name) }

	switch {
	case set("secret") && set("secret_wo"):
		return fmt.Errorf("auth: secret and secret_wo are mutually exclusive")
	case !set("secret") && !set("secret_wo"):
		return fmt.Errorf("auth: one of secret or secret_wo is required")
	case set("secret_wo") && !set("secret_wo_version"):
		return fmt.Errorf("auth: secret_wo requires secret_wo_version")
	}

	if !d.NewValueKnown("auth.0.type") {
		return nil
	}
	authType := d.Get("auth.0.type").(string)
	if authType == client.MCPAuthTypeBasic {
		if !set("username") {
			return fmt.Errorf("auth: username is required with type %q", authType)
		}
	} else if set("username") {
		return fmt.Errorf("auth: username is only valid with type %q", client.MCPAuthTypeBasic)
	}
	if authType == client.MCPAuthTypeHeader {
		if !set("name") {
			return fmt.Errorf("auth: name is required with type %q", authType)
		}
		if err := validateAuthHeaderName(d); err != nil {
			return err
		}
	} else if set("name") {
		return fmt.Errorf("auth: name is only valid with type %q", client.MCPAuthTypeHeader)
	}

	// OAuth sends its token in Authorization too, so only one of them can.
	if authType != client.MCPAuthTypeHeader && writeonly.IsSet(d, "client_id") {
		return fmt.Errorf("auth: type %q sets the Authorization header, which OAuth needs "+
			"for client_id; use type %q instead", authType, client.MCPAuthTypeHeader)
	}
	return nil
}

func validateAuthHeaderName(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("auth.0.name") {
		return nil
	}
	name := d.Get("auth.0.name").(string)
	if strings.EqualFold(name, "authorization") {
		return fmt.Errorf("auth: use type %q or %q to set the Authorization header",
			client.MCPAuthTypeBearer, client.MCPAuthTypeBasic)
	}
	// heimdall holds the auth header to the same rule as headers.
	if reason := reservedHeader(name); reason != "" {
		return fmt.Errorf("auth: %s", reason)
	}
	if !d.NewValueKnown("headers") {
		return nil
	}
	for header := range d.Get("headers").(map[string]any) {
		if strings.EqualFold(header, name) {
			return fmt.Errorf("auth: header %q is already set in headers", name)
		}
	}
	return nil
}

// reservedHeader says why a header name is one the gateway owns, or "" when
// it is free to set.
func reservedHeader(name string) string {
	lowered := strings.ToLower(name)
	if deniedHeaderNames[lowered] {
		return fmt.Sprintf("%q is set by the gateway and cannot be configured", name)
	}
	for _, prefix := range deniedHeaderPrefixes {
		if strings.HasPrefix(lowered, prefix) {
			return fmt.Sprintf("%q is reserved: no header may start with %q", name, prefix)
		}
	}
	return ""
}
