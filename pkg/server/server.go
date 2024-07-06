package server

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pavlov061356/http_based_file_storage/internal/helpers"
	"github.com/pavlov061356/http_based_file_storage/pkg/storage"
)

// TODO: additional hash check on POST with user provided hashing algs

type FileStorageServer interface {

	// SaveFile handles the HTTP POST request to save a file to the storage.
	// It saves the file to a temporary location, computes its hash, and saves it to the storage.
	// If an error occurs during the process, it returns an error 500 Internal Server Error.
	// If the file already exists in the storage, it returns a status code 200 OK.
	// If the file is successfully saved, it returns a status code 201 Created and the hash of the file.
	// Support checking user-provided hashes in headers on request in camel case
	// Hash checking supports MD5, SHA256, SHA512, SHA1 hashes
	SaveFile(c *gin.Context)

	// SendFile handles the HTTP GET request to retrieve a file from the storage.
	// It retrieves the file from the storage based on the provided hash and sends it back as the response.
	// If the file is not found in the storage, it returns an error 404 Not Found.
	// If the file is successfully retrieved, it returns a status code 200 OK and the file.
	// If an error occurs during the process, it returns an error 500 Internal Server Error.
	//
	// Parameters:
	// - c: The Gin context object for handling the HTTP request and response.
	SendFile(c *gin.Context)

	// DeleteFile handles the HTTP DELETE request to delete a file from the storage.
	// It checks if the file exists in the storage, and if so, deletes it.
	// Returns 200 OK if the file is successfully deleted or if file does not exists.
	// Returns an error 500 Internal Server Error if an internal error occurs.
	//
	// Parameters:
	// - c: The Gin context object for handling the HTTP request and response.
	DeleteFile(c *gin.Context)

	// StartServer starts the HTTP server.
	// It sets up the router and starts the server to listen for incoming requests.
	//
	// Returns:
	// - error: any error that occurs during startup
	StartServer()

	// setupRouter sets up the Gin router with the appropriate routes and handlers.
	// It returns a pointer to the configured Gin engine.
	//
	// Returns:
	// - *gin.Engine: the configured Gin engine
	setupRouter() *gin.Engine

	// RegisterPreSaveCallback registers a callback function to be executed before
	// a file is saved.
	//
	// The callback function takes two parameters: the hash of the file and the path
	// to the file. It must return an error if there was any problem executing
	// the callback.
	//
	// This function is thread-safe.
	RegisterPreSaveCallback(callback func(hash string, filePath string) error)

	// RegisterPOSTSaveCallback registers a callback function to be executed after
	// a file is successfully saved.
	//
	// The callback function takes two parameters: the hash of the saved file and
	// the path to the saved file. It must return an error if there was any problem
	// executing the callback.
	//
	// This function is thread-safe.
	RegisterPOSTSaveCallback(callback func(hash string, filePath string) error)

	// RegisterGETHandler registers a handler function for the GET method on the
	// specified path.
	//
	// Parameters:
	// - path: the path to register the handler on
	// - handler: the function to handle GET requests on the specified path
	RegisterGETHandler(path string, handler func(c *gin.Context))

	// RegisterPOSTHandler registers a handler function for the POST method on the
	// specified path.
	//
	// Parameters:
	// - path: the path to register the handler on
	// - handler: the function to handle POST requests on the specified path
	RegisterPOSTHandler(path string, handler func(c *gin.Context))

	// RegisterDELETEHandler registers a handler function for the DELETE method on the
	// specified path.
	//
	// Parameters:
	// - path: the path to register the handler on
	// - handler: the function to handle DELETE requests on the specified path
	RegisterDELETEHandler(path string, handler func(c *gin.Context))

	// AddMiddleware adds a middleware function to the Gin engine.
	//
	// Parameters:
	// - middleware: the function to add as middleware
	AddMiddleware(middleware gin.HandlerFunc)
}

type HTTPFileStorageServer struct {
	storer storage.Storer
	config *Config

	mux sync.Mutex

	engine *gin.Engine

	preSaveCallbacks []func(hash string, filePath string) error

	postSaveCallbacks []func(hash string, filePath string) error
}

type hash struct {
	Hash string `uri:"hash" binding:"required"`
}

// setupRouter sets up the Gin router with the appropriate routes and handlers.
// It returns a pointer to the configured Gin engine.
func (s *HTTPFileStorageServer) setupRouter() *gin.Engine {
	// Create a new Gin engine with the default middleware
	r := gin.Default()

	// Add the recovery middleware to handle panics
	r.Use(gin.Recovery())

	// Add routes and handlers
	// POST /file - SaveFile handler for saving files
	r.POST("/file", s.SaveFile)
	// GET /file/:hash - SendFile handler for retrieving files
	r.GET("/file/:hash", s.SendFile)
	// DELETE /file/:hash - DeleteFile handler for deleting files
	r.DELETE("/file/:hash", s.DeleteFile)

	// Return the configured Gin engine
	return r
}

// StartServer starts the HTTP server.
// It sets up the router and starts the server to listen for incoming requests.
func (s *HTTPFileStorageServer) StartServer() {
	// Set up the router
	r := s.setupRouter()

	// Create a new HTTP server
	server := &http.Server{
		// Set the address to listen on
		Addr: fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		// Set the handler to the router
		Handler: r,
		// Set the timeouts for the server
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	s.engine = r
	go func() {
		// Listen and serve
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			// Print the error if the server fails to start
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.

	<-ctx.Done()
	log.Println("timeout of 5 seconds.")

	log.Println("Server exiting")

}

// SaveFile handles the HTTP POST request to save a file to the storage.
// It saves the file to a temporary location, computes its hash, and saves it to the storage.
// If an error occurs during the process, it returns an error 500 Internal Server Error.
// If the file already exists in the storage, it returns a status code 200 OK.
// If the file is successfully saved, it returns a status code 201 Created and the hash of the file.
func (s *HTTPFileStorageServer) SaveFile(c *gin.Context) {

	waitCh := make(chan struct{})
	go func() {

		defer func() {
			// Recover from panic and return error 500 Internal Server Error
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
				c.AbortWithError(500, fmt.Errorf("internal server error"))
			}
		}()

		defer func() { waitCh <- struct{}{} }()

		// Log the request
		slog.Info("POST /file")

		// Get the form file
		formFile, err := c.FormFile("file")
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error getting file: %v", err))
			return
		}

		// Create a temporary file
		file, err := os.CreateTemp("", formFile.Filename)
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error creating temp file: %v", err))
			return
		}

		// Open the form file
		multipartFile, err := formFile.Open()
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error opening file: %v", err))
			return
		}

		// Copy the form file to the temporary file
		_, err = io.Copy(file, multipartFile)
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error saving file locally: %v", err))
			return
		}

		// Close the form file and temporary file
		defer func() {
			multipartFile.Close()
			file.Close()
			os.Remove(file.Name())
		}()

		hash := helpers.GetFileHash(sha256.New(), file)

		if hash == "" {
			c.AbortWithError(500, fmt.Errorf("error computing hash: %v", err))
			return
		}

		// Close the temporary file because it will be read in SaveFileFromTemp
		file.Close()

		err = checkHashFromRequest(file.Name(), c)

		if err != nil {
			c.AbortWithError(412, fmt.Errorf("error checking hash: %v", err))
			return
		}

		// Run all Pre-Save callbacks
		s.runCallbacks(&s.preSaveCallbacks, hash, file.Name())

		// Save the file to the storage
		err = s.storer.SaveFileFromTemp(hash, file.Name())

		// If the file already exists in the storage, return a status code 200 OK
		if errors.Is(err, os.ErrExist) {
			c.Status(200)
			return
		}

		// If an error occurs during saving, return an error 500 Internal Server Error
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error saving file: %v", err))
			return
		}

		// Run all Post-Save callbacks
		s.runCallbacks(&s.postSaveCallbacks, hash, file.Name())

		// Return the hash of the file
		c.JSON(201, gin.H{"hash": hash})
	}()

	<-waitCh
}

// SendFile handles the HTTP GET request to retrieve a file from the storage.
// It checks if the file exists in the storage, and if so, sends it to the client.
// If the file does not exist, it returns an error 404 Not Found.
// If an internal error occurs, it returns error 500 Internal Server Error.
func (s *HTTPFileStorageServer) SendFile(c *gin.Context) {
	waitCh := make(chan struct{})
	go func() {
		defer func() {
			// Recover from panic and return error 500 Internal Server Error
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
				c.AbortWithError(500, fmt.Errorf("internal server error"))
			}
		}()

		defer func() { waitCh <- struct{}{} }()

		// Bind URI parameters to hash struct
		var hash hash
		if err := c.ShouldBindUri(&hash); err != nil {
			// Return error 400 Bad Request if URI parameters cannot be bound
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		// Read file from storage
		filePath, err := s.storer.Read(hash.Hash)

		if errors.Is(err, os.ErrNotExist) {
			// Return error 404 Not Found if file does not exist
			c.AbortWithError(404, fmt.Errorf("file not found"))
			return
		} else if err != nil {
			// Return error 500 Internal Server Error if an internal error occurs
			c.AbortWithError(500, fmt.Errorf("error reading file: %v", err))
			return
		}

		file, err := os.Open(filePath)

		if err != nil {
			// Return error 500 Internal Server Error if an internal error occurs
			c.AbortWithError(500, fmt.Errorf("error opening file: %v", err))
			return
		}

		computedHash := helpers.GetFileHash(sha256.New(), file)

		file.Close()
		if computedHash == "" {
			c.AbortWithError(500, fmt.Errorf("error computing hash: %v", err))
			return
		}

		if hash.Hash != computedHash {
			// TODO обсудить варианты возврата ошибок
			// Return error 500 with text "File is corrupted" if hash does not match
			// Deletes file after that
			c.AbortWithError(500, fmt.Errorf("File is corrupted"))
			file.Close()
			os.Remove(filePath)
			return
		}

		// Send file to client
		c.File(filePath)

		// Delete file from temporary directory
		os.Remove(filePath)
	}()

	<-waitCh
}

// DeleteFile handles the HTTP DELETE request to delete a file from the storage.
// It checks if the file exists in the storage, and if so, deletes it.
// Returns an error 500 Internal Server Error if an internal error occurs.
func (s *HTTPFileStorageServer) DeleteFile(c *gin.Context) {

	waitCh := make(chan struct{})
	go func() {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
				c.AbortWithError(500, fmt.Errorf("internal server error"))
			}
		}()

		defer func() { waitCh <- struct{}{} }()

		var hash hash
		if err := c.ShouldBindUri(&hash); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		if s.storer == nil {
			c.AbortWithError(500, fmt.Errorf("storage not initialized"))
			return
		}

		err := s.storer.Delete(hash.Hash)

		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error deleting file: %v", err))
			return
		}
		c.Status(200)
	}()

	<-waitCh
}

// NewHTTPFileStorageServer creates a new HTTPFileStorageServer instance.
//
// Parameters:
// - storer: the storage.Storer implementation to use for file storage
// - config: the server configuration
//
// Returns:
// - FileStorageServer: the created HTTPFileStorageServer instance
// - error: any error that occurred during initialization
func NewHTTPFileStorageServer(storer storage.Storer, config *Config) (FileStorageServer, error) {
	// Check if storer is nil
	if storer == nil {
		return nil, fmt.Errorf("storer field is nil")
	}

	// Check if config is nil
	if config == nil {
		return nil, fmt.Errorf("config field is nil")
	}

	// Create and return a new HTTPFileStorageServer instance
	return &HTTPFileStorageServer{
		storer:            storer,
		config:            config,
		mux:               sync.Mutex{},
		preSaveCallbacks:  []func(hash string, filePath string) error{},
		postSaveCallbacks: []func(hash string, filePath string) error{},
	}, nil
}

// registerCallback appends a callback function to the given slice of callbacks.
//
// The callbacks slice is protected by a mutex to ensure thread safety.
//
// Parameters:
// - callbacks: a pointer to a slice of callback functions
// - callback: the callback function to append to the slice
func (s *HTTPFileStorageServer) registerCallback(callbacks *[]func(hash string, filePath string) error, callback func(hash string, filePath string) error) {
	// Lock the mutex to prevent concurrent access to the callbacks slice
	s.mux.Lock()
	defer s.mux.Unlock()

	// Append the callback function to the callbacks slice
	*callbacks = append(*callbacks, callback)
}

// RegisterPreSaveCallback registers a callback function to be executed before
// a file is saved.
//
// The callback function takes two parameters: the hash of the file and the path
// to the file. It must return an error if there was any problem executing
// the callback.
//
// This function is thread-safe.
func (s *HTTPFileStorageServer) RegisterPreSaveCallback(callback func(hash string, filePath string) error) {
	s.registerCallback(&s.preSaveCallbacks, callback)
}

// RegisterPOSTSaveCallback registers a callback function to be executed after
// a file is successfully saved.
//
// The callback function takes two parameters: the hash of the saved file and
// the path to the saved file. It must return an error if there was any problem
// executing the callback.
//
// This function is thread-safe.
//
// Parameters:
// - callback: the callback function to register.
func (s *HTTPFileStorageServer) RegisterPOSTSaveCallback(callback func(hash string, filePath string) error) {
	s.registerCallback(&s.postSaveCallbacks, callback)
}

// runCallbacks runs all the callbacks in the given slice with the provided
// hash and file path.
//
// The callbacks slice is not modified.
//
// Parameters:
// - callbacks: a pointer to a slice of callbacks.
// - hash: the hash of the file.
// - filePath: the path of the file.
func (s *HTTPFileStorageServer) runCallbacks(callbacks *[]func(hash string, filePath string) error, hash string, filePath string) {

	// Lock the mutex to prevent concurrent access.
	s.mux.Lock()
	defer s.mux.Unlock()

	// Iterate over each callback in the slice.
	for _, callback := range *callbacks {
		// Call the callback with the provided hash and file path.
		callback(hash, filePath)
	}

}

// RegisterGETHandler registers a handler function for the GET method on the
// specified path.
//
// Parameters:
// - path: the path to register the handler on
// - handler: the function to handle GET requests on the specified path
func (s *HTTPFileStorageServer) RegisterGETHandler(path string, handler func(c *gin.Context)) {
	s.engine.GET(path, handler)
}

// RegisterDELETEHandler registers a handler function for the DELETE method on the
// specified path.
//
// Parameters:
// - path: the path to register the handler on
// - handler: the function to handle DELETE requests on the specified path
func (s *HTTPFileStorageServer) RegisterDELETEHandler(path string, handler func(c *gin.Context)) {
	s.engine.DELETE(path, handler)
}

// RegisterPOSTHandler registers a handler function for the POST method on the
// specified path.
//
// Parameters:
// - path: the path to register the handler on
// - handler: the function to handle POST requests on the specified path
func (s *HTTPFileStorageServer) RegisterPOSTHandler(path string, handler func(c *gin.Context)) {
	s.engine.POST(path, handler)
}

// AddMiddleware adds a middleware function to the Gin engine.
//
// Parameters:
// - middleware: the function to add as middleware
func (s *HTTPFileStorageServer) AddMiddleware(middleware gin.HandlerFunc) {
	s.engine.Use(middleware)
}

// checkHashFromRequest checks the hash of the file from the request headers.
//
// # Supports MD5 and SHA256, SHA512, SHA1 hashes
//
// Parameters:
// - filePath: the path of the file.
// - c: the gin context.
//
// Returns:
// - an error if the hash does not match, nil otherwise.
func checkHashFromRequest(filePath string, c *gin.Context) error {
	// Get the MD5 hash from the request header.
	md5Header := c.GetHeader("MD5")

	// Open the file.
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Check the MD5 hash.
	if md5Header != "" {
		hash := helpers.GetFileHash(md5.New(), file)
		if md5Header != hash {
			return fmt.Errorf("MD5 hash does not match")
		}
	}

	// Check the SHA256 hash.
	sha256Header := c.GetHeader("SHA256")
	if sha256Header != "" {
		hash := helpers.GetFileHash(sha256.New(), file)
		if sha256Header != hash {
			return fmt.Errorf("SHA256 hash does not match")
		}
	}

	// Check the SHA512 hash.
	sha512Header := c.GetHeader("SHA512")
	if sha512Header != "" {
		hash := helpers.GetFileHash(sha512.New(), file)
		if sha512Header != hash {
			return fmt.Errorf("SHA512 hash does not match")
		}
	}

	// Check the SHA1 hash.
	sha1Header := c.GetHeader("SHA1")
	if sha1Header != "" {
		hash := helpers.GetFileHash(sha1.New(), file)
		if sha1Header != hash {
			return fmt.Errorf("SHA1 hash does not match")
		}
	}

	return nil
}
