package repositories

import "testing"

func TestSafeOrderBy(t *testing.T) {
	allowed := map[string]string{
		"title":      "i.title",
		"created_at": "i.created_at",
	}
	const def = "i.updated_at"

	cases := []struct {
		col, dir string
		want     string
	}{
		{"title", "ASC", "i.title ASC"},
		{"title", "asc", "i.title ASC"},
		{"created_at", "DESC", "i.created_at DESC"},
		{"created_at", "", "i.created_at DESC"},              // default dir
		{"nonexistent", "ASC", "i.updated_at ASC"},           // junk col -> default col
		{"; DROP TABLE items;--", "ASC", "i.updated_at ASC"}, // injection -> default col
		{"title", "; DELETE", "i.title DESC"},                // junk dir -> DESC
		{"TITLE", "asc", "i.title ASC"},                      // case-insensitive col
	}
	for _, c := range cases {
		if got := safeOrderBy(c.col, c.dir, allowed, def); got != c.want {
			t.Errorf("safeOrderBy(%q,%q) = %q, want %q", c.col, c.dir, got, c.want)
		}
	}
}

func TestEscapeLike(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc", "abc"},
		{"100%", `100\%`},
		{"a_b", `a\_b`},
		{`a\b`, `a\\b`},
		{"%_\\", `\%\_\\`},
	}
	for _, c := range cases {
		if got := escapeLike(c.in); got != c.want {
			t.Errorf("escapeLike(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
