package types

// Ternary returns trueVal if condition is true, falseVal otherwise.
func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
