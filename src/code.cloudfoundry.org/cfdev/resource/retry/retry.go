package retry

func Retry(fn func() error, shouldRetry func(error) bool) error {
	for {
		err := fn()
		if err == nil {
			return nil
		} else if !shouldRetry(err) {
			return err
		}
	}
}

type retryable struct {
	err error
}

func (e *retryable) Error() string {
	return e.err.Error()
}

func WrapAsRetryable(err error) error {
	return &retryable{err}
}

func Retryable(retries int) func(error) bool {
	counter := 0
	return func(e error) bool {
		counter++
		_, isRetry := e.(*retryable)
		return isRetry && counter < retries
	}
}
