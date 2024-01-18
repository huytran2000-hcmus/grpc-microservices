package config

import (
	"log"
	"os"
	"strconv"
)

func GetEnv() string {
	return getEnvironmentValue("ENV")
}

func GetDataSourceURL() string {
	return getEnvironmentValue("DATA_SOURCE_URL")
}

func GetApplicationPort() int {
	portStr := getEnvironmentValue("APPLICATION_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("port: %s is invalid", portStr)
	}

	return port
}

func GetOTLPEndpoint() string {
	return getEnvironmentValue("OTLP_ENDPOINT")
}

func getEnvironmentValue(key string) string {
	return os.Getenv(key)
}
