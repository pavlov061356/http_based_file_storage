package server

import "github.com/gin-gonic/gin"

type FileStorageServer interface {
	// TODO: add methods below:
	// 1. AddFile
	// Get hash of file and save it to storage
	// OPtional: check hashes sent by client and compute hashes of uploaded file
	// if they are not equal, return error 412 Precondition Failed
	// Successful code is 201 Created
	// If an internal error occurs, return error 500 Internal Server Error

	UploadFile(c *gin.Context)
	// 2. GetFile
	// Get file from storage and return it
	// Check hash sent by client and compute hash of downloaded file
	// if they isn't any file with given hash return error 404 Not Found
	// Successful code is 200 OK
	// If an internal error occurs, return error 500 Internal Server Error
	GetFile(c *gin.Context)
	// 3. DeleteFile
	// Delete file from storage
	// if they isn't any file with given hash return error 404 Not Found
	// TODO: discuss 404
	// Successful code is 200 OK
	// If an internal error occurs, return error 500 Internal Server Error
	DeleteFile(c *gin.Context)
}
