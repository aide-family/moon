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

// GetOr returns the value pointed to by the given pointer.
func GetOr[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}

// GetOrZero returns the value pointed to by the given pointer.
func GetOrZero[T any](p *T) (T, bool) {
	if p != nil {
		return *p, true
	}
	var zero T
	return zero, false
}
