package services

import (
	"database/sql"
	"strings"
	"testing"

	"qvarkk/kvault/internal/domain"
)

func TestMeetsAutotagContentThreshold(t *testing.T) {
	long := strings.Repeat("x", 200)

	cases := []struct {
		name  string
		item  domain.Item
		count int
		want  bool
	}{
		{
			name:  "short text item fails",
			item:  domain.Item{Title: "hi", Content: sql.NullString{String: "tiny", Valid: true}},
			count: 3,
			want:  false, // 6 chars < 150
		},
		{
			name:  "long content passes",
			item:  domain.Item{Title: "title", Content: sql.NullString{String: long, Valid: true}},
			count: 3,
			want:  true, // 205 >= 150
		},
		{
			name:  "url item with only extracted content passes",
			item:  domain.Item{Title: "t", ExtractedContent: sql.NullString{String: long, Valid: true}},
			count: 3,
			want:  true, // extracted content must count
		},
		{
			name:  "url item with empty content/extracted fails",
			item:  domain.Item{Title: "t"},
			count: 1,
			want:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := meetsAutotagContentThreshold(&c.item, c.count); got != c.want {
				t.Errorf("meetsAutotagContentThreshold = %v, want %v", got, c.want)
			}
		})
	}
}
