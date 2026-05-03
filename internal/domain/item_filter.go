package domain

type ListItemFilter struct {
	UserID string
	Type   string
	TagIDs []string
	QueryFilter
	PaginationFilter
	SortFilter
}
