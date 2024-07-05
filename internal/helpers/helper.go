package helpers

import "fmt"

// GetFilePath generates the file path for a given hash value.
//
// hash: The hash value of the file.
// basePath: The base path where the file will be stored.
// Returns the file path as a string.
func GetFilePath(hash string, basePath string) string {
	// The file path is constructed by concatenating the base path,
	// the first two characters of the hash value, and the complete hash value.
	return fmt.Sprintf("%s/store/%s/%s", basePath, hash[0:2], hash)
}
