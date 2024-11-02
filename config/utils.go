package config

import "os"

func IsDevMode() bool {
	return os.Getenv("DEV_MODE") == "true"
}
