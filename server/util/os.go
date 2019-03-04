// Package util implements utility methods
package util

import (
	"fmt"
	"log"
	"os"
)

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

func createFileWithContent(filename, content string) error {

	// Create a file on disk
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("error while creating file: %s", err.Error())
		return fmt.Errorf("error while creating file: %s", err.Error())
	}
	defer file.Close()

	// Open the file to write
	file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("error while opening file: %s", err.Error())
		return fmt.Errorf("error while opening file: %s", err.Error())
	}

	// Write the Markdown doc to disk
	_, err = file.Write([]byte(content))
	if err != nil {
		log.Printf("error while writing Markdown to disk: %s", err.Error())
		return fmt.Errorf("error while writing Markdown to disk: %s", err.Error())
	}

	return nil
}
