package helpers

import "os"

func GetEnv() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "production"
	}
	return env
}
