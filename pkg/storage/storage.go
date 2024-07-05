package storage

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/pavlov061356/http_based_file_storage/internal/helpers"
)

const maxBufferSize = 1024

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

	saveFile(hash string, data []byte) error

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

	// TODO: could be made configurable
	bufferSize int
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
		basePath:   basePath,
		muxMap:     make(map[string]*sync.Mutex),
		bufferSize: maxBufferSize,
	}, nil
}

// Exists checks if a file with the given hash exists in the storage.
//
// hash: the hash of the file to check.
//
// Returns a boolean indicating if the file exists and an error if there was any.
func (s *Storage) Exists(hash string) (bool, error) {
	// Get the file path for the given hash
	filePath := helpers.GetFilePath(s.basePath, hash)

	// Check if the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		// If the file doesn't exist or there was an error while checking the file, return false and the error
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	// If the file exists, return true and no error
	return true, nil
}

// SaveFileFromTemp saves a file to the storage.
//
// hash: the hash of the file to save
// tmpFilePath: the path to the temporary file to save
//
// Returns an error if there was any
func (s *Storage) SaveFileFromTemp(hash string, tmpFilePath string) error {
	mux := createMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

	filePath := helpers.GetFilePath(s.basePath, hash)
	// Lock the mutex to prevent concurrent access to the file
	mux.Lock()
	defer mux.Unlock()

	defer deleteMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

	// Save the file by renaming the temporary file
	err := os.Rename(tmpFilePath, filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s *Storage) saveFile(hash string, data []byte) error {
	// Lock the mutex map to prevent concurrent access

	filePath := helpers.GetFilePath(s.basePath, hash)
	hashedFilePath := helpers.GetFileParentPath(s.basePath, hash)
	err := os.MkdirAll(hashedFilePath, os.ModePerm)
	if err != nil {
		return err
	}

	mux := createMutexMapEntry(&s.muxMapLock, s.muxMap, hash)
	mux.Lock()
	defer mux.Unlock()

	defer deleteMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

	err = os.WriteFile(filePath, data, 0644)
	if err != nil && !os.IsNotExist(err) {
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

	mux := createMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

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
		return "", os.ErrExist
	}
	mux.Lock()

	file, err := os.Open(filePath)

	if err != nil {
		// If there was an error while opening the file, return the error
		return "", err
	}
	defer file.Close()
	stat, err := file.Stat()

	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), hash)

	if err != nil {
		return "", err
	}
	tempFilePath := filepath.Join(tempDir, hash)

	temFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	defer temFile.Close()

	// Read the file and write it to the temporary file
	// The buffer size is computed for each file to check if file size is lower than the max buffer size
	// to avoid getting buffer filled like this: [bytes, ... 0, 0, 0, 0, ...]
	bufferSize := int(math.Min(float64(s.bufferSize), float64(stat.Size())))
	buffer := make([]byte, bufferSize)
	bufferedReader := bufio.NewReader(file)

	for {
		_, err := bufferedReader.Read(buffer)
		if err != nil {
			break
		}
		temFile.Write(buffer)
	}

	mux.Unlock()

	defer deleteMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

	// Return the path to the temporary file
	return tempFilePath, nil
}

// Delete deletes a file from the storage.
//
// hash: the hash of the file to delete
//
// Returns an error if there was any
func (s *Storage) Delete(hash string) error {
	mux := createMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

	// Get the file path for the given hash
	filePath := helpers.GetFilePath(s.basePath, hash)

	// Lock the mutex to prevent concurrent access to the file
	mux.Lock()
	defer mux.Unlock()
	defer deleteMutexMapEntry(&s.muxMapLock, s.muxMap, hash)

	// Delete the file
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
func deleteMutexMapEntry(muxMapLock *sync.Mutex, muxMap map[string]*sync.Mutex, hash string) {
	muxMapLock.Lock()

	// Delete the mutex associated with the hash
	delete(muxMap, hash)

	muxMapLock.Unlock()
}

func createMutexMapEntry(muxMapLock *sync.Mutex, muxMap map[string]*sync.Mutex, hash string) *sync.Mutex {
	muxMapLock.Lock()

	// Get the mutex associated with the hash
	mux, ok := muxMap[hash]
	if !ok {
		// If the mutex doesn't exist, create it
		mux = &sync.Mutex{}
		muxMap[hash] = mux
	}

	muxMapLock.Unlock()
	return mux
}
