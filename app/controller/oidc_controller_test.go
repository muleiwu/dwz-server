package controller

import (
	"net/url"
	"strings"
	"testing"
)

func TestAppendHashAwareQuery_HashMode(t *testing.T) {
	got := appendHashAwareQuery("/admin/#/auth/oidc-redirect", map[string]string{
		"token":      "abc.def.ghi",
		"expires_at": "1776960656",
	})
	// 期望:hash 后加 query,而非真 query。
	if !strings.HasPrefix(got, "/admin/#/auth/oidc-redirect?") {
		t.Fatalf("expected hash-embedded query, got %q", got)
	}
	// 解析 hash query 确认两个 key 都在。
	hashQuery := got[strings.Index(got, "?")+1:]
	vals, err := url.ParseQuery(hashQuery)
	if err != nil {
		t.Fatalf("parse hash query: %v", err)
	}
	if vals.Get("token") != "abc.def.ghi" {
		t.Errorf("token missing: %v", vals)
	}
	if vals.Get("expires_at") != "1776960656" {
		t.Errorf("expires_at missing: %v", vals)
	}
}

func TestAppendHashAwareQuery_HashWithExistingQuery(t *testing.T) {
	got := appendHashAwareQuery("/admin/#/profile?tab=bindings", map[string]string{
		"oidc_bind": "ok",
	})
	if !strings.Contains(got, "tab=bindings") {
		t.Fatalf("pre-existing query dropped: %q", got)
	}
	if !strings.Contains(got, "oidc_bind=ok") {
		t.Fatalf("new param missing: %q", got)
	}
	// 不应该出现真 query。
	before, _, _ := strings.Cut(got, "#")
	if strings.Contains(before, "?") {
		t.Fatalf("query leaked to pre-hash segment: %q", got)
	}
}

func TestAppendHashAwareQuery_NoHash(t *testing.T) {
	got := appendHashAwareQuery("/welcome", map[string]string{"x": "1"})
	if got != "/welcome?x=1" {
		t.Errorf("expected /welcome?x=1, got %q", got)
	}
}

func TestAppendHashAwareQuery_Empty(t *testing.T) {
	if got := appendHashAwareQuery("", map[string]string{"x": "1"}); got != "" {
		t.Errorf("expected passthrough for empty target, got %q", got)
	}
	if got := appendHashAwareQuery("/x", nil); got != "/x" {
		t.Errorf("expected passthrough for nil params, got %q", got)
	}
}

func TestSanitizeReturnTo(t *testing.T) {
	cases := map[string]string{
		"/admin/#/profile":        "/admin/#/profile",
		"":                        "",
		"  /ok  ":                 "/ok",
		"//evil.com":              "",
		"https://evil.com":        "",
		"javascript:alert(1)":     "",
		"/admin/#/auth/oidc":      "/admin/#/auth/oidc",
	}
	for input, want := range cases {
		if got := sanitizeReturnTo(input); got != want {
			t.Errorf("sanitizeReturnTo(%q) = %q, want %q", input, got, want)
		}
	}
}
