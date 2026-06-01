package repositories

import "testing"

func TestBuildTsQuery(t *testing.T) {
	cases := []struct{ in, want string }{
		{"foo bar", "foo:* & bar:*"},
		{"hello", "hello:*"},
		{"", ""},
		{"   ", ""},
		{"foo, bar!", "foo:* & bar:*"},
		{"foo !!! bar", "foo:* & bar:*"},
		{"multi-word", "multi-word:*"},
		{"abc123", "abc123:*"},
	}
	for _, c := range cases {
		if got := buildTsQuery(c.in); got != c.want {
			t.Errorf("buildTsQuery(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
