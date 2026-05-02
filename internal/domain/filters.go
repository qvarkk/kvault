package domain

type QueryFilter struct {
	Query string
}

type PaginationFilter struct {
	Page     int
	PageSize int
}

type SortFilter struct {
	Direction string
	Column    string
}
