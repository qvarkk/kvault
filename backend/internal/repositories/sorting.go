package repositories

import "strings"

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

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
