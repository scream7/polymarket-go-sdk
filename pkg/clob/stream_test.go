package clob

import (
	"context"
	"testing"
)

func TestStreamData(t *testing.T) {
	fetch := func(ctx context.Context, cursor string) ([]int, string, error) {
		switch cursor {
		case InitialCursor:
			return []int{1, 2}, "NEXT", nil
		case "NEXT":
			return []int{3}, EndCursor, nil
		default:
			return nil, EndCursor, nil
		}
	}

	var got []int
	for res := range StreamData(context.Background(), fetch) {
		if res.Err != nil {
			t.Fatalf("unexpected error: %v", res.Err)
		}
		got = append(got, res.Item)
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("unexpected items: %v", got)
	}
}
