package web

type DescendingSortingParams struct {
	Direction string `form:"order_by,default=DESC" binding:"oneof=DESC ASC"`
}

type AscendingSortingParams struct {
	Direction string `form:"order_by,default=ASC" binding:"oneof=DESC ASC"`
}

type ItemSortingParams struct {
	DescendingSortingParams
	Column string `form:"sort_by,default=updated_at" binding:"oneof=title created_at updated_at"`
}

type FileSortingParams struct {
	DescendingSortingParams
	Column string `form:"sort_by,default=created_at" binding:"oneof=original_name size created_at"`
}

type StopwordSortingParams struct {
	AscendingSortingParams
	Column string `form:"sort_by,default=source" binding:"oneof=word source updated_at"`
}
