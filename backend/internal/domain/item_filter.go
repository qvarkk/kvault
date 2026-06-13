package domain

type ListItemFilter struct {
	UserID string
	TagIDs []string
	QueryFilter
	PaginationFilter
	SortFilter
}
