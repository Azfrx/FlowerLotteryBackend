package controller

import (
	"flower-lottery-backend/model"
	"testing"
)

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
