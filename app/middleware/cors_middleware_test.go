package middleware

import "testing"

func TestDefaultCorsConfigAllowsWorkspaceHeader(t *testing.T) {
	config := DefaultCorsConfig()
	for _, header := range config.AllowHeaders {
		if header == HeaderWorkspaceID {
			return
		}
	}
	t.Fatalf("default CORS allow headers must include %s: %#v", HeaderWorkspaceID, config.AllowHeaders)
}
