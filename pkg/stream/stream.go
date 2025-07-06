package stream

type Stream[T any] []T

func (s Stream[T]) list() []T {
	return s
}

func (s Stream[T]) Filter(f func(T) bool) Stream[T] {
	valid := Stream[T]{}

	for i := range s {
		if f(s[i]) {
			valid = append(valid.list(), s[i])
		}
	}

	return valid
}

func (s Stream[T]) ForEach(f func(T)) {
	for i := range s {
		f(s[i])
	}
}

func (s Stream[U]) Map(f func(U) any) Stream[any] {
	mapped := Stream[any]{}

	for i := range s {
		mapped = append(mapped, f(s[i]))
	}

	return mapped
}
