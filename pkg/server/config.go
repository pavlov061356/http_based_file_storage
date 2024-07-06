package server

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config represents the server configuration.
type Config struct {
	// Host is the server host address.
	Host string `json:"host"`
	// Port is the server port number.
	Port int `json:"port"`
	// StoragePath is the path to the storage directory.
	StoragePath string `json:"storage_path"`
}

// ReadConfigFromEnv reads the server configuration from the environment variables.
//
// It loads the environment variables from the .env file and sets default values if
// the variables are not set.
//
// Returns:
// - *Config: the server configuration
func ReadConfigFromEnv() *Config {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("WARNING: err while loading env file: %v\n", err)
	}

	// Get the host address from the environment variable, default to "localhost"
	host, exists := os.LookupEnv("HOST")
	if !exists {
		host = "localhost"
	}

	// Get the port number from the environment variable, default to 8080
	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "8080"
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		parsedPort = 8080
	}

	// Get the storage path from the environment variable, default to "/tmp"
	storagePath, exists := os.LookupEnv("STORAGE_PATH")
	if !exists {
		storagePath = "/tmp"
	}

	// Create and return the server configuration
	return &Config{
		Host:        host,
		Port:        parsedPort,
		StoragePath: storagePath,
	}
}
