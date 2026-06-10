package domain

type ListTagFilter struct {
	UserID string
	QueryFilter
	PaginationFilter
	SortFilter
}
