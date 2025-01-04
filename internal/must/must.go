package must

func Return[T any](value T, err error) T {
	PanicOnError(err)
	return value
}

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func IgnoreError(err error) {}
