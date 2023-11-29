package config

import (
	"log"
	"os"
	"strconv"
)

func GetEnv() string {
	return getEnvironmentVarible("ENV")
}

func GetDataSourceURL() string {
	return getEnvironmentVarible("DATA_SOURCE_URL")
}

func GetApplicationPort() int {
	portStr := getEnvironmentVarible("APPLICATION_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("port: %s is invalid", portStr)
	}

	return port
}

func getEnvironmentVarible(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment varible is missing", key)
	}

	return val
}
