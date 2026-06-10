package services

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractMainText(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string
	}{
		{
			"adjacent elements stay separated",
			"<body><p>foo</p><p>bar</p></body>",
			"foo bar",
		},
		{
			"prefers article over body noise",
			"<body><nav>menu junk</nav><article><p>real content</p></article></body>",
			"real content",
		},
		{
			"prefers main when no article",
			"<body><header>skip me</header><main><p>main text</p></main></body>",
			"main text",
		},
		{
			"strips script style noscript",
			"<body><script>var x=1;</script><style>.a{}</style><noscript>nojs</noscript><p>visible</p></body>",
			"visible",
		},
		{
			"collapses whitespace",
			"<body><p>  hello   world  </p></body>",
			"hello world",
		},
	}
	for _, c := range cases {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(c.html))
		if err != nil {
			t.Fatalf("%s: parse: %v", c.name, err)
		}
		if got := extractMainText(doc); got != c.want {
			t.Errorf("%s: extractMainText() = %q, want %q", c.name, got, c.want)
		}
	}
}
