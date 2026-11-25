package scout

import (
	"os"
	"strings"
)

// childEnvAllowlist is what bundled tools need to run. Everything else in our
// environment (GITHUB_TOKEN, *_API_KEY, webhook secrets) stays out of reach of
// tools that parse untrusted repositories.
// ponytail: allowlist; add a name here if a bundled tool needs more.
var childEnvAllowlist = map[string]bool{
	"PATH": true, "HOME": true, "USER": true, "TMPDIR": true, "LANG": true, "LC_ALL": true, "TZ": true,
	"HTTP_PROXY": true, "HTTPS_PROXY": true, "NO_PROXY": true, "http_proxy": true, "https_proxy": true, "no_proxy": true,
	"SSL_CERT_FILE": true, "SSL_CERT_DIR": true,
	"TRIVY_CACHE_DIR": true,
}

// MinimalEnv returns the allowlisted part of the current environment plus extra
// (KEY=VALUE entries), for child processes that handle untrusted input.
func MinimalEnv(extra ...string) []string {
	var env []string
	for _, kv := range os.Environ() {
		if k, _, _ := strings.Cut(kv, "="); childEnvAllowlist[k] {
			env = append(env, kv)
		}
	}
	return append(env, extra...)
}
