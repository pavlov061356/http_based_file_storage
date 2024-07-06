package main

import (
	"github.com/pavlov061356/http_based_file_storage/pkg/server"
	"github.com/pavlov061356/http_based_file_storage/pkg/storage"
)

// main is the entry point of the application.
//
// It reads the configuration from environment variables, creates a new storage
// and a new HTTP file storage server, and starts the server.
func main() {
	// Read the configuration from environment variables.
	config := server.ReadConfigFromEnv()

	// Create a new storage.
	storage, err := storage.NewStorage(config.StoragePath)
	if err != nil {
		// Panic if an error occurred while creating the storage.
		panic(err)
	}

	// Create a new HTTP file storage server.
	server, err := server.NewHTTPFileStorageServer(storage, config)
	if err != nil {
		// Panic if an error occurred while creating the server.
		panic(err)
	}

	// Start the server.
	server.StartServer()
}
