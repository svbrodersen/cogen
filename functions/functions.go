package functions

func head[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[0], true
}

func tail[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[len(slice)-1], true
}

func o[T any](slice1 []T, slice2 []T) []T {
	slice1 = append(slice1, slice2...)
	return slice1
}

func list[T any](a ...T) []T {
	return a
}

func CallFunction[T any](name string, args ...T) {

}
