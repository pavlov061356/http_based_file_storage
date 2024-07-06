package helpers

import (
	"encoding/base64"
	"hash"
	"io"
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
	return filepath.Join(basePath, "store", hash[:2], hash)
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

// GetFileHash calculates the hash of a file and returns it as a base64 encoded string.
//
// hash: The hash function to use for calculating the hash of the file.
// file: The file to calculate the hash of.
// Returns the base64 encoded hash of the file as a string.
func GetFileHash(hash hash.Hash, file io.Reader) string {
	// Copy the contents of the file to the hash function.
	_, err := io.Copy(hash, file)
	if err != nil {
		// If there was an error copying the file, return an empty string.
		return ""
	}

	// Get the sum of the hash function.
	sum := hash.Sum(nil)

	// Encode the sum as a base64 string.
	base64Hash := base64.URLEncoding.EncodeToString(sum)

	// Return the base64 encoded hash.
	return base64Hash
}
