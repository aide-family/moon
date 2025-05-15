package pointer

// Of returns a pointer to the given value.
func Of[T any](v T) *T {
	return &v
}

// Get returns the value pointed to by the given pointer.
func Get[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}
