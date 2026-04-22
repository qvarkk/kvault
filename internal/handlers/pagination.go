package handlers

type PaginationParams struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

type PaginatedResponse struct {
	Data     any `json:"data"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func toPaginatedResponse(data any, total int, page int, page_size int) PaginatedResponse {
	return PaginatedResponse{
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: page_size,
	}
}
