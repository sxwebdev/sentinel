package utils

// Pointer
func Pointer[T any](v T) *T {
	return &v
}
