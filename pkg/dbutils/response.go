package dbutils

type FindResponseWithCount[T any] struct {
	Items []T    `json:"items"`
	Count uint32 `json:"count"`
}

func NewFindResponseWithCount[T any](items []T, count uint32) FindResponseWithCount[T] {
	if items == nil {
		items = []T{}
	}

	return FindResponseWithCount[T]{
		Items: items,
		Count: count,
	}
}
