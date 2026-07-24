package dto

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestUpdateShortLinkRequestCampaignIDTriState(t *testing.T) {
	t.Run("omitted keeps field unset", func(t *testing.T) {
		var req UpdateShortLinkRequest
		if err := json.Unmarshal([]byte(`{"title":"unchanged campaign"}`), &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req.CampaignID.Set || req.CampaignID.Value != nil {
			t.Fatalf("omitted campaign_id should stay unset: %+v", req.CampaignID)
		}
	})

	t.Run("null explicitly clears", func(t *testing.T) {
		var req UpdateShortLinkRequest
		if err := json.Unmarshal([]byte(`{"campaign_id":null}`), &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if !req.CampaignID.Set || req.CampaignID.Value != nil {
			t.Fatalf("null campaign_id should be explicitly set to nil: %+v", req.CampaignID)
		}
	})

	t.Run("positive integer explicitly sets", func(t *testing.T) {
		var req UpdateShortLinkRequest
		if err := json.Unmarshal([]byte(`{"campaign_id":42}`), &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if !req.CampaignID.Set || req.CampaignID.Value == nil || *req.CampaignID.Value != 42 {
			t.Fatalf("positive campaign_id was not preserved: %+v", req.CampaignID)
		}
	})

	for _, body := range []string{
		`{"campaign_id":0}`,
		`{"campaign_id":-1}`,
		`{"campaign_id":1.5}`,
		`{"campaign_id":"42"}`,
	} {
		t.Run("rejects "+body, func(t *testing.T) {
			var req UpdateShortLinkRequest
			if err := json.Unmarshal([]byte(body), &req); err == nil {
				t.Fatalf("expected invalid campaign_id to fail: %s", body)
			}
		})
	}
}

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
