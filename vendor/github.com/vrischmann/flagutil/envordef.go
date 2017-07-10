package flagutil

import "os"

// EnvOrDefault returns the environment variable value at `envName` if not empty,
// otherwise it returns `defaultVal`
func EnvOrDefault(envName, defaultVal string) string {
	if e := os.Getenv(envName); e != "" {
		return e
	}
	return defaultVal
}
