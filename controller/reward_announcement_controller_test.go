package controller

import (
	"bytes"
	"encoding/json"
	"flower-lottery-backend/service"
	"strings"
	"testing"
	"time"
)

func TestWriteRewardAnnouncementEvent(t *testing.T) {
	announcement := service.RewardAnnouncement{
		ID:         42,
		Nickname:   "花花",
		ItemCode:   "1205251",
		RewardName: "月亮游记戒指",
		GrantedAt:  time.Date(2026, time.July, 14, 18, 30, 0, 0, time.UTC),
	}
	var output bytes.Buffer
	if err := writeRewardAnnouncementEvent(&output, announcement); err != nil {
		t.Fatal(err)
	}

	stream := output.String()
	if !strings.HasPrefix(stream, "id: 42\nevent: reward-announcement\ndata: ") {
		t.Fatalf("unexpected SSE prefix: %q", stream)
	}
	if !strings.HasSuffix(stream, "\n\n") {
		t.Fatalf("SSE event must end with a blank line: %q", stream)
	}
	dataLine := strings.Split(stream, "\n")[2]
	var payload service.RewardAnnouncement
	if err := json.Unmarshal([]byte(strings.TrimPrefix(dataLine, "data: ")), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ID != announcement.ID || payload.Nickname != announcement.Nickname || payload.RewardName != announcement.RewardName {
		t.Fatalf("payload = %+v, want %+v", payload, announcement)
	}
}

func TestParseRewardAnnouncementLastEventID(t *testing.T) {
	if got := parseRewardAnnouncementLastEventID(" 18 ", "9"); got != 18 {
		t.Fatalf("header cursor = %d, want 18", got)
	}
	if got := parseRewardAnnouncementLastEventID("invalid", "9"); got != 9 {
		t.Fatalf("query cursor = %d, want 9", got)
	}
	if got := parseRewardAnnouncementLastEventID("", "invalid"); got != 0 {
		t.Fatalf("invalid cursor = %d, want 0", got)
	}
}
