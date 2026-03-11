package utils

func If[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func Ptr[T any](v T) *T {
	return &v
}
