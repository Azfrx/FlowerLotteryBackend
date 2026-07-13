package service

import (
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
