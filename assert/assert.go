package assert

import (
	"homepage/panic"
)

func OK(err error) {
	if err != nil {
		panic.Go(err)
	}
}

func T[T any](r T, err error) T {
	if err != nil {
		panic.Go(err)
	}
	return r
}
