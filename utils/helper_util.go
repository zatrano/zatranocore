package utils

import (
	"fmt"
	"os"
)

func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	result, err := fmt.Sscanf(value, "%d")
	if err != nil {
		return defaultValue
	}
	return result
}
