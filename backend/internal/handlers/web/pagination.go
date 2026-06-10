package web

type PaginationParams struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

type PaginatedResponse[T any] struct {
	Data     []T `json:"data"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func toPaginatedResponse[T any](data []T, total int, page int, page_size int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: page_size,
	}
}
