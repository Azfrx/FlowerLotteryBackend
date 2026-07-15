package controller

import (
	"flower-lottery-backend/model"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRewardItemGoingOffline(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus uint8
		nextStatus    uint8
		want          bool
	}{
		{name: "enabled to disabled", currentStatus: 1, nextStatus: 0, want: true},
		{name: "enabled stays enabled", currentStatus: 1, nextStatus: 1, want: false},
		{name: "disabled stays disabled", currentStatus: 0, nextStatus: 0, want: false},
		{name: "disabled to enabled", currentStatus: 0, nextStatus: 1, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewardItemGoingOffline(tt.currentStatus, tt.nextStatus); got != tt.want {
				t.Fatalf("rewardItemGoingOffline(%d, %d) = %v, want %v", tt.currentStatus, tt.nextStatus, got, tt.want)
			}
		})
	}
}

func TestOnlinePrizePoolRewardReferencesUsesPublishedEnabledPool(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "test:test@tcp(127.0.0.1:3306)/test?parseTime=true",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	var count int64
	query := onlinePrizePoolRewardReferences(db, 42).Count(&count)
	if query.Error != nil {
		t.Fatalf("build reference query: %v", query.Error)
	}
	sql := query.Statement.SQL.String()
	for _, fragment := range []string{
		"JOIN prize_pool_versions AS v ON v.id=pr.version_id AND v.status=1",
		"JOIN prize_pools AS p ON p.id=v.prize_pool_id AND p.status=1 AND p.deleted_at IS NULL",
		"pr.reward_item_id=?",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("reference query %q does not contain %q", sql, fragment)
		}
	}
}

func TestValidAdminResourceURL(t *testing.T) {
	for _, value := range []string{"", "/uploads/reward.png", "https://example.com/reward.png", "http://example.com/reward.svga"} {
		if !validAdminResourceURL(value) {
			t.Fatalf("expected %q to be accepted", value)
		}
	}
	for _, value := range []string{"javascript:alert(1)", "example.com/reward.png", "ftp://example.com/reward.png"} {
		if validAdminResourceURL(value) {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestNormalizeAndValidateRewardItem(t *testing.T) {
	input := adminRewardItemInput{
		ItemCode: " 1001 ", Name: " 金币 ", ItemType: "coin",
		ImageURL: "https://example.com/coin.png", Status: 1,
	}
	if !normalizeAndValidateRewardItem(&input, true) {
		t.Fatal("expected valid reward item")
	}
	if input.ItemCode != "1001" || input.Name != "金币" {
		t.Fatalf("expected values to be normalized, got %#v", input)
	}
	input.ItemType = "unknown"
	if normalizeAndValidateRewardItem(&input, true) {
		t.Fatal("expected unknown reward type to be rejected")
	}
}

func TestNormalizeAndValidateExchangeOption(t *testing.T) {
	input := adminExchangeOptionInput{
		ActivityID: 1, PetalAmount: 50, CoinCost: 3000,
		SortNo: 2, Status: 1, Remark: "  常用档位  ",
	}
	if !normalizeAndValidateExchangeOption(&input) {
		t.Fatal("expected valid exchange option")
	}
	if input.Remark != "常用档位" {
		t.Fatalf("expected remark to be normalized, got %q", input.Remark)
	}
	input.CoinCost = 0
	if normalizeAndValidateExchangeOption(&input) {
		t.Fatal("expected zero coin cost to be rejected")
	}
	input.CoinCost = 3000
	input.SortNo = -1
	if normalizeAndValidateExchangeOption(&input) {
		t.Fatal("expected negative sort number to be rejected")
	}
}

func TestAdminCoinValueForPetalCost(t *testing.T) {
	tests := []struct {
		name      string
		petalCost uint64
		want      uint64
		valid     bool
	}{
		{name: "zero", petalCost: 0, valid: false},
		{name: "day pool", petalCost: 5, want: 300, valid: true},
		{name: "night pool", petalCost: 100, want: 6000, valid: true},
		{name: "overflow", petalCost: ^uint64(0)/adminCoinValuePerPetal + 1, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := adminCoinValueForPetalCost(tt.petalCost)
			if got != tt.want || valid != tt.valid {
				t.Fatalf("adminCoinValueForPetalCost(%d) = (%d, %v), want (%d, %v)", tt.petalCost, got, valid, tt.want, tt.valid)
			}
		})
	}
}

func TestValidateAdminActivityContent(t *testing.T) {
	content := model.ActivityContent{
		Instructions: model.ActivityInstructionsContent{
			Title: "活动说明",
			Sections: []model.ActivityInstructionSection{{
				Title:      "活动时间",
				Paragraphs: [][]model.ActivityTextSegment{{{Text: "活动进行中"}}},
			}},
			ProbabilityLink: model.ActivityLink{Text: "概率公示", URL: "https://example.com/probability"},
		},
		GameGuides: model.ActivityGameGuidesContent{
			Day:   [][]model.ActivityGuideNode{{{Type: "common", Content: "白昼攻略"}}},
			Night: [][]model.ActivityGuideNode{{{Type: "text", Content: "星夜攻略"}}},
		},
		RankingCustomization: model.ActivityInstructionsContent{
			Title: "冠名说明",
			Sections: []model.ActivityInstructionSection{{
				Title:      "头像框冠名",
				Paragraphs: [][]model.ActivityTextSegment{{{Text: "冠名规则"}}},
			}},
		},
	}
	if err := validateAdminActivityContent(content); err != nil {
		t.Fatalf("expected valid activity content: %v", err)
	}
	content.GameGuides.Day[0][0].Type = "unknown"
	if err := validateAdminActivityContent(content); err == nil {
		t.Fatal("expected invalid guide node type to be rejected")
	}
}
