package storage

import (
	"fmt"
	"os"
	"sync"

	"github.com/pavlov061356/http_based_file_storage/internal/helpers"
)

// Storer is an interface that defines the methods for file storage
type Storer interface {
	// Exists checks if a file with the given hash exists
	//
	// hash: the hash of the file to check
	//
	// Returns a boolean indicating if the file exists and an error if there was any
	Exists(hash string) (bool, error)

	// SaveFileFromTemp saves a file to the storage
	//
	// hash: the hash of the file to save
	// tmpFilePath: the path to the temporary file to save
	//
	// Returns an error if there was any
	SaveFileFromTemp(hash string, tmpFilePath string) error

	SaveFile(hash string, data []byte) error

	// Read reads a file from the storage
	//
	// hash: the hash of the file to read
	//
	// Returns the path to the file and an error if there was any
	Read(hash string) (string, error)

	// Delete deletes a file from the storage
	//
	// hash: the hash of the file to delete
	//
	// Returns an error if there was any
	Delete(hash string) error
}

// Storage represents a file storage system.
type Storage struct {
	// basePath is the base directory where the files are stored.
	basePath string

	// muxMap is a map of mutexes used to synchronize file access.
	// The key is the hash of the file, and the value is the mutex associated with that hash.
	muxMap map[string]*sync.Mutex

	// muxMapLock is a mutex used to synchronize access to the muxMap.
	muxMapLock sync.Mutex
}

// NewStorage creates a new instance of Storage with the specified base path.
//
// basePath: the base path where the files will be stored.
//
// Returns a pointer to a Storage instance and an error if there was any.
func NewStorage(basePath string) (Storer, error) {
	// Check if the base path exists
	_, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		// If it doesn't exist, create the directory
		err = os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		// If there was an error while checking the directory, return the error
		return nil, err
	}
	// Return a new Storage instance
	return &Storage{
		basePath: basePath,
		muxMap:   make(map[string]*sync.Mutex),
	}, nil
}

func (s *Storage) Exists(hash string) (bool, error) {
	filePath := helpers.GetFilePath(s.basePath, hash)

	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SaveFileFromTemp saves a file to the storage.
//
// hash: the hash of the file to save
// tmpFilePath: the path to the temporary file to save
//
// Returns an error if there was any
func (s *Storage) SaveFileFromTemp(hash string, tmpFilePath string) error {
	// Lock the mutex map to prevent concurrent access
	s.muxMapLock.Lock()

	// Get the mutex associated with the hash
	mux, ok := s.muxMap[hash]
	if !ok {
		// If the mutex doesn't exist, create it
		mux = &sync.Mutex{}
		s.muxMap[hash] = mux
	}

	// Unlock the mutex map
	s.muxMapLock.Unlock()

	filePath := helpers.GetFilePath(s.basePath, hash)
	// Lock the mutex to prevent concurrent access to the file
	mux.Lock()
	defer mux.Unlock()

	// Save the file by renaming the temporary file
	err := os.Rename(tmpFilePath, filePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveFile(hash string, data []byte) error {
	filePath := helpers.GetFilePath(s.basePath, hash)
	err := os.WriteFile(filePath, data, 0511)
	if err != nil {
		return err
	}
	return nil
}

// Read reads a file from the storage and writes it to a temporary file.
//
// hash: the hash of the file to read
//
// Returns the path to the file and an error if there was any
func (s *Storage) Read(hash string) (string, error) {
	// Get the file path for the given hash
	filePath := helpers.GetFilePath(s.basePath, hash)

	// Check if the file exists
	exists, err := s.Exists(hash)
	if err != nil {
		// If there was an error while checking the file, return the error
		return "", err
	}

	// If the file doesn't exist, return nil
	if !exists {
		return "", nil
	}

	// Read the file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If there was an error while reading the file, return the error
		return "", err
	}

	// Create a temporary file with the same hash name
	tempFilePath := fmt.Sprintf("%s/%s", os.TempDir(), hash)
	err = os.WriteFile(tempFilePath, data, 0511)

	// If there was an error while writing the temporary file, return the error
	if err != nil {
		return "", err
	}
	// Return the path to the temporary file
	return filePath, nil
}

// Delete deletes a file from the storage.
//
// hash: the hash of the file to delete
//
// Returns an error if there was any
func (s *Storage) Delete(hash string) error {
	// Lock the mutex map to prevent concurrent access
	s.muxMapLock.Lock()

	// Get the mutex associated with the hash
	mux, ok := s.muxMap[hash]
	if !ok {
		// If the mutex doesn't exist, create it
		mux = &sync.Mutex{}
		s.muxMap[hash] = mux
	}

	// Unlock the mutex map
	s.muxMapLock.Unlock()

	// Get the file path for the given hash
	filePath := helpers.GetFilePath(s.basePath, hash)

	// Lock the mutex to prevent concurrent access to the file
	mux.Lock()
	defer mux.Unlock()

	// Delete the file
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
