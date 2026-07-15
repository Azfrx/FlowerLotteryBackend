package controller

import (
	"bytes"
	"encoding/json"
	"flower-lottery-backend/service"
	"strings"
	"testing"
	"time"
)

func TestWritePoolConfigEvent(t *testing.T) {
	update := service.PoolConfigUpdate{
		ID: 7, PoolID: 2, PoolCode: "night", PetalCostPerDraw: 120,
		CoinValuePerDraw: 7200, VersionNo: 4,
		PublishedAt: time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC),
	}
	var output bytes.Buffer
	if err := writePoolConfigEvent(&output, update); err != nil {
		t.Fatal(err)
	}

	stream := output.String()
	if !strings.HasPrefix(stream, "id: 7\nevent: pool-config-updated\ndata: ") {
		t.Fatalf("unexpected SSE prefix: %q", stream)
	}
	if !strings.HasSuffix(stream, "\n\n") {
		t.Fatalf("SSE event must end with a blank line: %q", stream)
	}
	dataLine := strings.Split(stream, "\n")[2]
	var payload service.PoolConfigUpdate
	if err := json.Unmarshal([]byte(strings.TrimPrefix(dataLine, "data: ")), &payload); err != nil {
		t.Fatal(err)
	}
	if payload != update {
		t.Fatalf("payload = %+v, want %+v", payload, update)
	}
}
