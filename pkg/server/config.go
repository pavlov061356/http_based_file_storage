package server

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	StoragePath string `json:"storage_path"`
}

func ReadConfigFromEnv() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("WARNING: err while loading env file: %v\n", err)
	}

	host, exists := os.LookupEnv("host")

	if !exists {
		host = "localhost"
	}

	port, exists := os.LookupEnv("port")

	if !exists {
		port = "8080"
	}

	parsedPort, err := strconv.Atoi(port)

	if err != nil {
		parsedPort = 8080
	}

	storagePath, exists := os.LookupEnv("storage_path")

	if !exists {
		storagePath = "/tmp"
	}

	return &Config{
		Host:        host,
		Port:        parsedPort,
		StoragePath: storagePath,
	}
}
