package services

import (
	"database/sql"
	"strings"
)

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// stripNullBytes removes NUL (0x00) bytes, which PostgreSQL text columns reject
// ("invalid byte sequence for encoding UTF8: 0x00"). Extracted PDF/URL text can
// contain them; storing such text would fail the whole INSERT/UPDATE.
func stripNullBytes(s string) string {
	if !strings.ContainsRune(s, 0) {
		return s
	}
	return strings.ReplaceAll(s, "\x00", "")
}
