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

func TestBatchUpdateShortLinkStatusRequestAllowsFalseStatus(t *testing.T) {
	inactive := false
	req := BatchUpdateShortLinkStatusRequest{
		IDs:      []uint64{1, 2},
		IsActive: &inactive,
	}

	if err := binding.Validator.ValidateStruct(&req); err != nil {
		t.Fatalf("is_active=false should be accepted by binding validation: %v", err)
	}
}

func TestBatchShortLinkIDsRequestRejectsEmptyAndInvalidIDs(t *testing.T) {
	tests := []struct {
		name string
		req  BatchShortLinkIDsRequest
	}{
		{name: "empty IDs", req: BatchShortLinkIDsRequest{IDs: []uint64{}}},
		{name: "zero ID", req: BatchShortLinkIDsRequest{IDs: []uint64{1, 0}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := binding.Validator.ValidateStruct(&tt.req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
