package web

type ListResponse[T any] struct {
	Data []T `json:"data"`
}

func toListResponse[T any](data []T) ListResponse[T] {
	return ListResponse[T]{Data: data}
}
