package i18n

import (
	"strings"
)

const (
	LocaleEN = "en"
	LocaleRU = "ru"
	LocaleJA = "ja"
)

var supported = map[string]bool{LocaleEN: true, LocaleRU: true, LocaleJA: true}

func ParseLocale(header string) string {
	for part := range strings.SplitSeq(header, ",") {
		tag := strings.TrimSpace(part)
		tag = strings.SplitN(tag, ";", 2)[0]
		tag = strings.SplitN(strings.TrimSpace(tag), "-", 2)[0]
		tag = strings.ToLower(tag)
		if supported[tag] {
			return tag
		}
	}
	return LocaleEN
}
