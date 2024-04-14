package redis

func RetryRedis(err error) (retryCause string, outcome bool) {
	// fixme check for an actual error that can be retried and return true if so
	return err.Error(), true
}
