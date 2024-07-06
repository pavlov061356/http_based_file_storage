package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pavlov061356/http_based_file_storage/internal/helpers"
	"github.com/stretchr/testify/assert"
)

// TestStorageCreate tests the creation of a new Storage instance.
//
// It verifies that a new Storage instance can be created with the specified base path.
func TestStorageCreate(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")

	// Assert that there is no error while creating the Storage instance.
	assert.NoError(t, err)

	// Assert that the Storage instance is not nil.
	assert.NotNil(t, storage)
}

// TestStorageSaveFile tests the SaveFile method of the Storage.
//
// It verifies that the SaveFile method correctly saves a file to the storage.
// It tests saving a file with a hash that doesn't exist in the storage yet,
// and saving a file with a hash that already exists in the storage.
func TestStorageSaveFile(t *testing.T) {

	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")

	// Assert that there is no error while creating the Storage instance.
	assert.NoError(t, err, "Error while creating Storage instance")

	// Assert that the Storage instance is not nil.
	assert.NotNil(t, storage, "Storage instance is nil")

	// Save a file with a hash that doesn't exist in the storage yet.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err, "Error while saving file with a new hash")

	// Assert that the file was saved successfully.
	filePath := helpers.GetFilePath("/tmp", "hash")
	_, err = os.Stat(filePath)
	assert.NoError(t, err, "File was not saved successfully")

	// Save a file with a hash that already exists in the storage.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err, "Error while saving file with an existing hash")

	_, err = os.Stat(filePath)
	assert.NoError(t, err, "File was not saved successfully")
}

// TestConcurrentStorageSaveFile tests concurrent SaveFile calls on the Storage.
//
// It verifies that the SaveFile method correctly saves a file to the storage
// when called concurrently by multiple goroutines.
// The test creates 50 goroutines, each calling SaveFile with a fixed hash
// and data. It asserts that no error occurs during the SaveFile calls.
// Finally, it asserts that the file was saved successfully.
func TestConcurrentStorageSaveFile(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")

	// Assert that there is no error while creating the Storage instance.
	assert.NoError(t, err, "Error while creating Storage instance")

	// Assert that the Storage instance is not nil.
	assert.NotNil(t, storage, "Storage instance is nil")

	// Create a WaitGroup to wait for all goroutines to finish.
	wg := sync.WaitGroup{}

	// Add the number of goroutines to the WaitGroup.
	wg.Add(50)

	// Iterate 50 times and create a goroutine to call SaveFile.
	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			// Call SaveFile with a fixed hash and data.
			err := storage.saveFile("hash", []byte("data"))

			// Assert that no error occurs during the SaveFile call.
			assert.NoError(t, err)
		}()
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	// Get the file path.
	filePath := helpers.GetFilePath("/tmp", "hash")

	// Assert that the file was saved successfully.
	_, err = os.Stat(filePath)
	assert.NoError(t, err, "File was not saved successfully")
}

// TestStorageExists tests the Exists method of the Storage.
//
// It verifies that the Exists method correctly checks if a file exists in the storage.
// It tests saving a file with a hash that doesn't exist in the storage yet,
// and checking if the file exists using the Exists method.
// It also verifies that the file was saved successfully.
func TestStorageExists(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")
	assert.NoError(t, err, "Error while creating Storage instance")
	assert.NotNil(t, storage, "Storage instance is nil")

	// Save a file with a hash that doesn't exist in the storage yet.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err, "Error while saving file with a new hash")

	// Check if the file exists using the Exists method.
	exists, err := storage.Exists("hash")
	assert.NoError(t, err, "Error while checking if file exists")
	assert.True(t, exists, "File does not exist")

	// Get the file path.
	filePath := helpers.GetFilePath("/tmp", "hash")

	// Verify that the file was saved successfully.
	_, err = os.Stat(filePath)
	assert.NoError(t, err, "File was not saved successfully")
}

// TestStorageDelete tests the Delete method of the Storage.
//
// It verifies that the Delete method correctly deletes a file from the storage.
// It tests saving a file with a hash, deleting it, and checking if the file exists using the Exists method.
func TestStorageDelete(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Save a file with a hash.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err)

	// Delete the file.
	err = storage.Delete("hash")
	assert.NoError(t, err)

	// Check if the file exists.
	exists, err := storage.Exists("hash")
	assert.NoError(t, err)

	// Assert that the file was deleted successfully.
	assert.False(t, exists, "File does not exist")
}

// TestDeleteOnNonExistentFile tests the Delete method of the Storage
// when trying to delete a file that does not exist.
//
// It verifies that the Delete method does not return an error when
// trying to delete a file that does not exist.
func TestDeleteOnNonExistentFile(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Create a temporary directory in the base path.
	tempDir, err := os.MkdirTemp("/tmp", "test")
	assert.NoError(t, err)

	// Create a file path in the temporary directory.
	testFileName := filepath.Join(tempDir, "test.txt")

	// Try to delete the file.
	err = storage.Delete(testFileName)

	// Assert that there is no error while deleting the file.
	assert.NoError(t, err)
}

// TestStorageExistsOnDeletedFile tests the Exists method of the Storage
// when trying to check the existence of a file that has been deleted.
//
// It verifies that the Exists method correctly returns false when
// trying to check the existence of a file that has been deleted.
func TestStorageExistsOnDeletedFile(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Save a file with a hash.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err)

	// Delete the file.
	err = os.Remove(helpers.GetFilePath("/tmp", "hash"))
	assert.NoError(t, err)

	// Check if the file exists.
	exists, err := storage.Exists("hash")
	assert.NoError(t, err)

	// Assert that the file does not exist.
	assert.False(t, exists, "File does not exist")
}

// TestStorageRead tests the Read method of the Storage.
//
// It verifies that the Read method correctly reads a file from the storage.
// It tests saving a file with a hash, reading it using the Read method,
// and asserting that the file content is as expected.
func TestStorageRead(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Save a file with a hash.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err)

	// Read the file using the Read method.
	filePathFromStorage, err := storage.Read("hash")
	assert.NoError(t, err)

	// Verify that the file was saved successfully.
	_, err = os.Stat(filePathFromStorage)
	assert.NoError(t, err, "File was not saved successfully")

	// Read the file content.
	fileContent, err := os.ReadFile(filePathFromStorage)

	// Assert that the file was read successfully.
	assert.NoError(t, err, "File was not saved successfully")

	// Assert that the file content is as expected.
	assert.Equal(t, "data", string(fileContent))
}

// TestStorageReadOnDeletedFile tests the Read method of the Storage
// when trying to read a file that has been deleted.
//
// It verifies that the Read method correctly returns an error when
// trying to read a file that has been deleted.
func TestStorageReadOnDeletedFile(t *testing.T) {
	// Create a new Storage instance with the specified base path.
	storage, err := NewStorage("/tmp")
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Save a file with a hash.
	err = storage.saveFile("hash", []byte("data"))
	assert.NoError(t, err)

	// Delete the file.
	err = os.Remove(helpers.GetFilePath("/tmp", "hash"))
	assert.NoError(t, err)

	// Try to read the file using the Read method.
	_, err = storage.Read("hash")

	// Assert that the Read method returns an error.
	assert.Error(t, err)
}
