package model

import "testing"

func TestOIDCProvider_IsExclusive(t *testing.T) {
	cases := []struct {
		name     string
		provider *OIDCProvider
		want     bool
	}{
		{"nil", nil, false},
		{"disabled-noexclusive", &OIDCProvider{Enabled: 0, Exclusive: 0}, false},
		{"disabled-exclusive", &OIDCProvider{Enabled: 0, Exclusive: 1}, false},
		{"enabled-noexclusive", &OIDCProvider{Enabled: 1, Exclusive: 0}, false},
		{"enabled-exclusive", &OIDCProvider{Enabled: 1, Exclusive: 1}, true},
	}
	for _, c := range cases {
		if got := c.provider.IsExclusive(); got != c.want {
			t.Errorf("%s: IsExclusive() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestOIDCProvider_IsEnabled(t *testing.T) {
	if (*OIDCProvider)(nil).IsEnabled() {
		t.Error("nil provider should not be enabled")
	}
	if (&OIDCProvider{Enabled: 0}).IsEnabled() {
		t.Error("Enabled=0 should report disabled")
	}
	if !(&OIDCProvider{Enabled: 1}).IsEnabled() {
		t.Error("Enabled=1 should report enabled")
	}
}
