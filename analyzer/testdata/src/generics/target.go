package main

type myType[T any] struct {
	value T
}

func (m *myType[T]) set(v T) {
	m.value = v
}
