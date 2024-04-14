package helpers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {

	var retryError = errors.New("retryable error")

	var retryableErr = func(error) (retryCause string, outcome bool) {
		return retryError.Error(), true
	}

	mainFn := func(attempts int, isRetryableFns []IsRetryableFn, expectedError error) (int, error) {
		var count int
		err := Retry(func(attempt int, lastRetryCause string) error {
			count++
			return expectedError
		}, attempts, time.Second, isRetryableFns...)
		return count, err
	}

	for _, c := range []struct {
		name           string
		attempts       int
		retry          func(attempts int, isRetryableFns []IsRetryableFn, expectedError error) (int, error)
		isRetryableFns []IsRetryableFn
		expectedErr    error
	}{
		{
			name:           "retry 3 times",
			attempts:       3,
			retry:          mainFn,
			isRetryableFns: []IsRetryableFn{retryableErr},
			expectedErr:    retryError,
		},
		{
			name:           "retry 1 time",
			attempts:       1,
			retry:          mainFn,
			isRetryableFns: []IsRetryableFn{},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			count, err := c.retry(c.attempts, c.isRetryableFns, c.expectedErr)
			require.Equal(t, c.expectedErr, err)
			require.Equal(t, c.attempts, count)
		})
	}

}
