package domain

type ListItemFilter struct {
	UserID string
	Type   string
	QueryFilter
	PaginationFilter
	SortFilter
}
