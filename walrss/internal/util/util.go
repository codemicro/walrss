package util

func Contains[T comparable](x []T, y T) bool {
	for _, z := range x {
		if z == y {
			return true
		}
	}
	return false
}
