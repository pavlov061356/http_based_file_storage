package helpers

import (
	"path/filepath"
)

// GetFilePath generates the file path for a given hash value.
//
// hash: The hash value of the file.
// basePath: The base path where the file will be stored.
// Returns the file path as a string.
func GetFilePath(basePath string, hash string) string {
	// The file path is constructed by concatenating the base path,
	// the first two characters of the hash value, and the complete hash value.
	return filepath.Join(basePath, "store", hash[0:2], hash)
}

// GetFileParentPath generates the parent directory path for a given hash value.
//
// basePath: The base path where the file will be stored.
// hash: The hash value of the file.
// Returns the parent directory path as a string.
func GetFileParentPath(basePath string, hash string) string {
	// The parent directory path is constructed by concatenating the base path,
	// the first two characters of the hash value.
	// This is the parent directory where the file with the given hash will be stored.
	return filepath.Join(basePath, "store", hash[0:2])
}
