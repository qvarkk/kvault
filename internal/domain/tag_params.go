package domain

type ListTagParams struct {
	UserID    string
	Query     string
	Page      int
	PageSize  int
	Direction string
	Column    string
}
