package controller

import (
	"strings"
	"testing"

	"cnb.cool/mliev/dwz/dwz-server/v2/app/dto"
	"github.com/gin-gonic/gin/binding"
)

func TestBindErrorMessageShortLinkURL(t *testing.T) {
	req := dto.CreateShortLinkRequest{OriginalURL: "example.com"}
	err := binding.Validator.ValidateStruct(&req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	got := bindErrorMessage(err)
	want := "请求参数错误: 原始 URL 格式无效，请输入完整 URL，例如 https://example.com"
	if got != want {
		t.Fatalf("bindErrorMessage() = %q, want %q", got, want)
	}
}

func TestBindErrorMessageFallbackURL(t *testing.T) {
	req := dto.CreateShortLinkRequest{
		OriginalURL: "https://example.com",
		FallbackURL: "fallback",
	}
	err := binding.Validator.ValidateStruct(&req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	got := bindErrorMessage(err)
	if !strings.Contains(got, "兜底地址 URL 格式无效") || !strings.Contains(got, "或留空") {
		t.Fatalf("fallback URL message should be actionable, got %q", got)
	}
}
