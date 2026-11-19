package credutil

import (
	"fmt"
	"strings"
)

// MySQLSSLModes are the ssl_mode values the API takes for mysql and mariadb
// credentials, hyphenated as midgard's MySQLSSLMode and MariaDBSSLMode are.
var MySQLSSLModes = []string{
	"disabled", "preferred", "required", "verify-ca", "verify-identity",
}

// MySQLSSLModeDesc describes ssl_mode for a mysql or mariadb credential;
// server names the database in the first sentence.
func MySQLSSLModeDesc(server string) string {
	return strings.Join([]string{
		fmt.Sprintf("The SSL mode for the %s connection (default: preferred).", server),
		"`disabled`: no TLS.",
		"`preferred`: TLS when the server offers it, without verifying its certificate.",
		"`required`: TLS, without verifying the certificate.",
		"`verify-ca`: TLS, with the certificate verified against `ssl_ca`, or the system CAs without it.",
		"`verify-ca` does not check the host name; `verify-identity` also requires the certificate to name the host.",
		"Older hush-am versions verify the certificate under `required`, accept any certificate under `verify-ca`, and need the server to offer TLS under `preferred`.",
	}, " ")
}
