package domain

type ListItemParams struct {
	UserID    string
	Query     string
	Type      string
	Page      int
	PageSize  int
	Direction string
	Column    string
}
