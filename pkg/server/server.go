package server

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pavlov061356/http_based_file_storage/pkg/storage"
)

// TODO: graceful shutdown with context
// TODO: check file hash on GET
// TODO: add pre- and post-callbacks on saving file
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

	StartServer()

	setupRouter() *gin.Engine
}

type HTTPFileStorageServer struct {
	storer storage.Storer
	config *Config
}

type hash struct {
	Hash string `uri:"hash" binding:"required"`
}

func (s *HTTPFileStorageServer) setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.POST("/file", s.SaveFile)
	r.GET("/file/:hash", s.SendFile)
	r.DELETE("/file/:hash", s.DeleteFile)
	return r
}

func (s *HTTPFileStorageServer) StartServer() {
	r := s.setupRouter()
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error occured on https server: %v", err)
		return
	}
}

func (s *HTTPFileStorageServer) SaveFile(c *gin.Context) {

	waitCh := make(chan struct{})
	go func() {

		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
				c.AbortWithError(500, fmt.Errorf("internal server error"))
			}
		}()

		defer func() { waitCh <- struct{}{} }()

		slog.Info("POST /file")
		formFile, err := c.FormFile("file")
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error getting file: %v", err))
			return
		}

		file, err := os.CreateTemp("", formFile.Filename)

		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error creating temp file: %v", err))
			return
		}
		multipartFile, err := formFile.Open()

		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error opening file: %v", err))
			return
		}
		_, err = io.Copy(file, multipartFile)

		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error saving file locally: %v", err))
			return
		}
		defer func() {
			multipartFile.Close()
			file.Close()
			os.Remove(file.Name())
		}()

		h := sha256.New()
		_, err = io.Copy(h, file)

		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error computing hash: %v", err))
			return
		}
		hash := fmt.Sprintf("%x", h.Sum(nil))
		// Closing file because it will be read in SaveFileFromTemp
		file.Close()

		err = s.storer.SaveFileFromTemp(hash, file.Name())

		if errors.Is(err, os.ErrExist) {
			c.Status(200)
			return
		}
		if err != nil {
			c.AbortWithError(500, fmt.Errorf("error saving file: %v", err))
			return
		}
		c.JSON(201, gin.H{"hash": hash})
	}()

	<-waitCh
}

func (s *HTTPFileStorageServer) SendFile(c *gin.Context) {
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
		slog.Info("hash", slog.String("hash", hash.Hash))
		filePath, err := s.storer.Read(hash.Hash)

		slog.Info("filePath", slog.String("filePath", filePath))

		if errors.Is(err, os.ErrNotExist) {
			c.AbortWithError(404, fmt.Errorf("file not found"))
			return
		} else if err != nil {
			c.AbortWithError(500, fmt.Errorf("error reading file: %v", err))
			return
		}
		//Send file to client
		c.File(filePath)

		//Delete file from temp dir
		os.Remove(filePath)
	}()

	<-waitCh
}

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

func NewHTTPFileStorageServer(storer storage.Storer, config *Config) (FileStorageServer, error) {
	if storer == nil {
		return nil, fmt.Errorf("storer field is nil")
	}

	if config == nil {
		return nil, fmt.Errorf("config field is nil")
	}
	return &HTTPFileStorageServer{
		storer,
		config,
	}, nil
}
