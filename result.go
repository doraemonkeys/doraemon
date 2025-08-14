package doraemon

type Pair[T1, T2 any] struct {
	First  T1
	Second T2
}

type Result[T any] struct {
	Err   error
	Value T
}

func Ok[T any](value T) Result[T] {
	return Result[T]{
		Err:   nil,
		Value: value,
	}
}

func Err[T any](err error) Result[T] {
	return Result[T]{
		Err: err,
	}
}

func (r Result[T]) IsOk() bool {
	return r.Err == nil
}

func (r Result[T]) IsErr() bool {
	return r.Err != nil
}

func (r Result[T]) Unwrap() T {
	if r.IsErr() {
		panic(r.Err)
	}
	return r.Value
}

func (r Result[T]) UnwrapErr() error {
	if r.IsOk() {
		panic("Result is ok")
	}
	return r.Err
}

func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.IsOk() {
		return r.Value
	}
	return defaultValue
}

func (r Result[T]) UnwrapOrElse(f func(error) T) T {
	if r.IsOk() {
		return r.Value
	}
	return f(r.Err)
}

func (r Result[T]) GoResult() (T, error) {
	return r.Value, r.Err
}
