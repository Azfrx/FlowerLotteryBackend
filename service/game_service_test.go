package service

import (
	"flower-lottery-backend/model"
	"reflect"
	"testing"
)

func TestChestUnlockThresholds(t *testing.T) {
	tests := []struct {
		name   string
		before uint8
		after  uint8
		want   []uint8
	}{
		{name: "does not reach a chest", before: 4, after: 5, want: []uint8{}},
		{name: "reaches the first chest", before: 5, after: 6, want: []uint8{6}},
		{name: "reaches the second chest", before: 11, after: 12, want: []uint8{12}},
		{name: "crosses two chest thresholds", before: 5, after: 12, want: []uint8{6, 12}},
		{name: "crosses a threshold without stopping on it", before: 11, after: 13, want: []uint8{12}},
		{name: "reaches the final chest", before: 17, after: 18, want: []uint8{18}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := chestUnlockThresholds(test.before, test.after)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("chestUnlockThresholds(%d, %d) = %v, want %v", test.before, test.after, got, test.want)
			}
		})
	}
}

func TestCurrentChestRewardRulesReplaceLegacyConfiguration(t *testing.T) {
	items := []model.RewardItem{
		{ID: 1, ItemCode: "1205251", Name: "月亮游记戒指"},
		{ID: 2, ItemCode: "1205470", Name: "爱意翩跹戒指"},
		{ID: 3, ItemCode: "1207751", Name: "真爱无敌戒指"},
	}
	configured := []model.ChestRewardRule{
		{ID: 10, RewardItemID: 3, Quantity: 1, Weight: 7, Status: 1, RewardItem: items[2]},
		{ID: 11, RewardItemID: 4, Quantity: 1, Weight: 1, Status: 1, RewardItem: model.RewardItem{ID: 4, ItemCode: "1203743", Name: "闪闪心蝶戒指"}},
		{ID: 12, RewardItemID: 5, Quantity: 1, Weight: 1, Status: 1, RewardItem: model.RewardItem{ID: 5, ItemCode: "1207244", Name: "烟花之恋戒指"}},
	}

	rules := currentChestRewardRules(8, 2, configured, items)
	if len(rules) != 3 {
		t.Fatalf("currentChestRewardRules() returned %d rules, want 3", len(rules))
	}

	gotCodes := []string{
		rules[0].RewardItem.ItemCode,
		rules[1].RewardItem.ItemCode,
		rules[2].RewardItem.ItemCode,
	}
	if !reflect.DeepEqual(gotCodes, chestRewardItemCodes) {
		t.Fatalf("currentChestRewardRules() codes = %v, want %v", gotCodes, chestRewardItemCodes)
	}
	if rules[2].ID != 10 || rules[2].Weight != 7 {
		t.Fatalf("current valid rule was not preserved: %+v", rules[2])
	}
	if rules[0].ActivityID != 8 || rules[0].ChestNo != 2 || rules[0].Weight != 1 {
		t.Fatalf("missing current rule did not receive the expected fallback: %+v", rules[0])
	}
}

func TestSelectedChestCandidateIgnoresLegacyReward(t *testing.T) {
	candidates := []model.UserChestCandidate{
		{ID: 1, Selected: 1, RewardItem: model.RewardItem{ItemCode: "1207244"}},
		{ID: 2, Selected: 1, RewardItem: model.RewardItem{ItemCode: "1205470"}},
	}

	selected := selectedChestCandidate(candidates)
	if selected == nil || selected.ID != 2 {
		t.Fatalf("selectedChestCandidate() = %+v, want current reward candidate 2", selected)
	}
}
