package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestActivityContentSeedContainsOperatorCopy(t *testing.T) {
	content := activityContentSeed()
	if len(content.Instructions.Sections) != 4 {
		t.Fatalf("expected four instruction sections, got %d", len(content.Instructions.Sections))
	}
	if len(content.GameGuides.Day) != 2 || len(content.GameGuides.Night) != 2 {
		t.Fatalf("unexpected game guide configuration: %#v", content.GameGuides)
	}
	if len(content.NewRingWelfare.SelectionSegments) == 0 || len(content.NewRingWelfare.FirstPublishSegments) == 0 {
		t.Fatal("new ring welfare copy must be included in the activity configuration")
	}
	if len(content.RankingCustomization.Sections) != 2 {
		t.Fatalf("expected two ranking customization sections, got %d", len(content.RankingCustomization.Sections))
	}
	if _, err := json.Marshal(content); err != nil {
		t.Fatalf("activity content must be JSON serializable: %v", err)
	}
	encoded, err := json.Marshal(content.Instructions)
	if err != nil {
		t.Fatalf("activity instructions must be JSON serializable: %v", err)
	}
	copyText := string(encoded)
	if !strings.Contains(copyText, "月亮游记戒指、爱意翩跹戒指或真爱无敌戒指自选权之一") {
		t.Fatal("activity instructions must describe the configured ring chest rewards")
	}
	if strings.Contains(copyText, "真爱无敌戒指自选权、闪闪心蝶戒指或烟花之恋戒指之一") {
		t.Fatal("activity instructions still describe the previous ring chest rewards")
	}
}
