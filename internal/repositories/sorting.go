package repositories

import "strings"

// safeOrderBy builds an "<column> <direction>" clause from a whitelist, so a
// column name can never be attacker-controlled even if a caller forgets to
// validate it. col is looked up in allowed (case-insensitive); unknown columns
// fall back to def. dir is normalized to ASC/DESC (default DESC).
func safeOrderBy(col, dir string, allowed map[string]string, def string) string {
	column, ok := allowed[strings.ToLower(strings.TrimSpace(col))]
	if !ok {
		column = def
	}

	direction := "DESC"
	if strings.EqualFold(strings.TrimSpace(dir), "ASC") {
		direction = "ASC"
	}

	return column + " " + direction
}

// escapeLike escapes the LIKE/ILIKE wildcards so user input is matched
// literally. Must be paired with `ESCAPE '\'` in the query.
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
