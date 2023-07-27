package internal

type Iterator[T any] struct {
	index  int
	values []T
}

func NewIterator[T any](values []T) *Iterator[T] {
	return &Iterator[T]{values: values}
}

func (i *Iterator[T]) HasNext() bool {
	if i.index < len(i.values) {
		return true
	}
	return false
}

func (i *Iterator[T]) HasPrevious() bool {
	if i.index > 0 {
		return true
	}
	return false
}

func (i *Iterator[T]) Next() T {
	if i.HasNext() {
		value := i.values[i.index]
		i.index++
		return value
	}
	panic("no more elements")
}

func (i *Iterator[T]) Previous() T {
	if i.HasPrevious() {
		i.index--
		return i.values[i.index]
	}
	panic("no more elements")
}
