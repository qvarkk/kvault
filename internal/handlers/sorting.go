package handlers

type SortingParams struct {
	Direction string `form:"order_by,default=DESC" binding:"oneof=DESC ASC"`
}

type ItemSortingParams struct {
	SortingParams
	Column string `form:"sort_by,default=updated_at" binding:"oneof=title created_at updated_at"`
}

type FileSortingParams struct {
	SortingParams
	Column string `form:"sort_by,default=created_at" binding:"oneof=original_name size created_at"`
}
