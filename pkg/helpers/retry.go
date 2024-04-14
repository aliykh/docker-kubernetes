package helpers

import (
	"log"
	"time"
)

type (
	RetryOpFn     func(attempt int, lastRetryCause string) error
	IsRetryableFn func(error) (retryCause string, outcome bool)
)

func Retry(retryOpFn RetryOpFn, maxAttempts int, maxBackOff time.Duration, isRetryableFns ...IsRetryableFn) (err error) {
	lastRetryCause := ""
	var backOff = time.Millisecond * 100
	for attempts := 0; attempts < maxAttempts; attempts++ {
		err = retryOpFn(attempts, lastRetryCause)
		if err == nil {
			return nil
		}

		isRetryable := false
		for _, fn := range isRetryableFns {
			cause, success := fn(err)
			if success {
				isRetryable = true
				lastRetryCause = cause
				break
			}
		}

		if !isRetryable {
			return err
		}

		log.Printf("retrying after %s, attempt: %d, cause: %s\n", backOff, attempts, lastRetryCause)
		time.Sleep(backOff)
		if backOff < maxBackOff {
			backOff *= 2
		}
	}

	return
}
