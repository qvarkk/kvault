package domain

type ListFileFilter struct {
	UserID   string
	MimeType string
	QueryFilter
	PaginationFilter
	SortFilter
}
