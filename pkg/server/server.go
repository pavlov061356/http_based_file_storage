package server

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pavlov061356/http_based_file_storage/pkg/storage"
)

// TODO: graceful shutdown with context
// TODO: check file hash on GET
// TODO: additional hash check on POST with user provided hashing algs

type FileStorageServer interface {
	// TODO: add methods below:
	// 1. AddFile
	// Get hash of file and save it to storage
	// OPtional: check hashes sent by client and compute hashes of uploaded file
	// if they are not equal, return error 412 Precondition Failed
	// Successful code is 201 Created
	// If an internal error occurs, return error 500 Internal Server Error

	SaveFile(c *gin.Context)
	// 2. SendFile
	// Get file from storage and return it
	// Check hash sent by client and compute hash of downloaded file
	// if they isn't any file with given hash return error 404 Not Found
	// Successful code is 200 OK
	// If an internal error occurs, return error 500 Internal Server Error
	SendFile(c *gin.Context)
	// 3. DeleteFile
	// Delete file from storage
	// if they isn't any file with given hash return error 404 Not Found
	// TODO: discuss 404
	// Successful code is 200 OK
	// If an internal error occurs, return error 500 Internal Server Error
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

	// RegisterPostSaveCallback registers a callback function to be executed after
	// a file is successfully saved.
	//
	// The callback function takes two parameters: the hash of the saved file and
	// the path to the saved file. It must return an error if there was any problem
	// executing the callback.
	//
	// This function is thread-safe.
	RegisterPostSaveCallback(callback func(hash string, filePath string) error)

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

	// Listen and serve
	err := server.ListenAndServe()
	if err != nil {
		// Print the error if the server fails to start
		fmt.Printf("error occurred on HTTP server: %v", err)
		return
	}
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

		// Compute the hash of the temporary file
		h := sha256.New()
		_, err = io.Copy(h, file)
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error computing hash: %v", err))
			return
		}
		hash := fmt.Sprintf("%x", h.Sum(nil))

		// Close the temporary file because it will be read in SaveFileFromTemp
		file.Close()

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

// RegisterPostSaveCallback registers a callback function to be executed after
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
func (s *HTTPFileStorageServer) RegisterPostSaveCallback(callback func(hash string, filePath string) error) {
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

func (s *HTTPFileStorageServer) RegisterGETHandler(path string, handler func(c *gin.Context)) {
	s.engine.GET(path, handler)
}

func (s *HTTPFileStorageServer) RegisterDELETEHandler(path string, handler func(c *gin.Context)) {
	s.engine.DELETE(path, handler)
}

func (s *HTTPFileStorageServer) RegisterPOSTHandler(path string, handler func(c *gin.Context)) {
	s.engine.POST(path, handler)
}

func (s *HTTPFileStorageServer) AddMiddleware(middleware gin.HandlerFunc) {
	s.engine.Use(middleware)
}
