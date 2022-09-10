package main

type myType[T any] struct {
	value T
}

func (m *myType[T]) set(v T) {
	m.value = v
}

func Map[T, R any](s []T, f func(T) R) []R {
	result := make([]R, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}
