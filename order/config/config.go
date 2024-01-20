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

func GetPaymentServiceURL() string {
	return getEnvironmentVarible("PAYMENT_SERVICE_URL")
}

func GetApplicationPort() int {
	portStr := getEnvironmentVarible("APPLICATION_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("port: %s is invalid", portStr)
	}

	return port
}

func GetOTLPEndpoint() string {
	return getEnvironmentVarible("OTLP_ENDPOINT")
}

func GetMetricAddress() string {
	return getEnvironmentVarible("METRIC_ADDRESS")
}

func getEnvironmentVarible(key string) string {
	val := os.Getenv(key)

	return val
}
