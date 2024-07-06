package main

import (
	"github.com/pavlov061356/http_based_file_storage/pkg/server"
	"github.com/pavlov061356/http_based_file_storage/pkg/storage"
)

func main() {
	config := server.ReadConfigFromEnv()
	storage, err := storage.NewStorage(config.StoragePath)
	server, err := server.NewHTTPFileStorageServer(storage, config)
	if err != nil {
		panic(err)
	}

	server.StartServer()
}
