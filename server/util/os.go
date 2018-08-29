// Package util implements utility methods
package util

import "os"

// GetEnvKey tries to get the specified key from the OS environment and returns either the
// value or the fallback that was provided
func GetEnvKey(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// HomeDir gets the homedir of the current user
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
