package service

import "testing"

func TestParseActivityContent(t *testing.T) {
	content, err := parseActivityContent([]byte(`{
		"instructions":{"title":"活动说明","sections":[],"probability_link":{"text":"概率公示","url":"https://example.com"}},
		"game_guides":{"day":[[{"type":"common","content":"白昼说明"}]],"night":[]},
		"new_ring_welfare":{"story_title":"戒指物语","story_lines":[],"value_text":"","selection_segments":[],"selection_names":[],"first_publish_segments":[],"first_publish_captions":[]}
	}`))
	if err != nil {
		t.Fatalf("parseActivityContent returned error: %v", err)
	}
	if content.Instructions.Title != "活动说明" {
		t.Fatalf("unexpected instructions title: %q", content.Instructions.Title)
	}
	if len(content.GameGuides.Day) != 1 || content.GameGuides.Day[0][0].Content != "白昼说明" {
		t.Fatalf("unexpected day guide: %#v", content.GameGuides.Day)
	}
}

func TestParseActivityContentAllowsEmptyConfiguration(t *testing.T) {
	content, err := parseActivityContent(nil)
	if err != nil {
		t.Fatalf("parseActivityContent returned error: %v", err)
	}
	if content.Instructions.Title != "" {
		t.Fatalf("expected empty content, got %#v", content)
	}
}

func TestParseActivityContentRejectsMalformedJSON(t *testing.T) {
	if _, err := parseActivityContent([]byte(`{"instructions":`)); err == nil {
		t.Fatal("expected malformed activity content to return an error")
	}
}
