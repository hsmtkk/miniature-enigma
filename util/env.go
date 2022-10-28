package util

import (
	"fmt"
	"os"
	"strconv"
)

func RequiredEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("you must define %s env var", key)
	}
	return val, nil
}

func GetPort() (int, error) {
	portStr, err := RequiredEnv("PORT")
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s as int; %w", portStr, err)
	}
	return port, nil
}
