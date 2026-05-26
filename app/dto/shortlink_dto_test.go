package dto

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestUpdateShortLinkRequestAllowsEmptyFallbackURL(t *testing.T) {
	empty := ""
	req := UpdateShortLinkRequest{FallbackURL: &empty}

	if err := binding.Validator.ValidateStruct(&req); err != nil {
		t.Fatalf("empty fallback_url should be accepted by binding validation: %v", err)
	}
}
