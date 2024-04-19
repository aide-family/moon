package ternary

// AOrB a or b
func AOrB(expr bool, a, b any) any {
	if expr {
		return a
	}
	return b
}
