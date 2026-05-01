package domain

type ListFileParams struct {
	UserID    string
	Query     string
	MimeType  string
	Page      int
	PageSize  int
	Direction string
	Column    string
}
