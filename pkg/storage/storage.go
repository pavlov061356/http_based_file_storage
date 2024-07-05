package storage

// Storer is an interface that defines the methods for file storage
type Storer interface {
	// Exists checks if a file with the given hash exists
	//
	// hash: the hash of the file to check
	//
	// Returns a boolean indicating if the file exists and an error if there was any
	Exists(hash string) (bool, error)

	// Save saves a file to the storage
	//
	// hash: the hash of the file to save
	// tmpFilePath: the path to the temporary file to save
	//
	// Returns an error if there was any
	Save(hash string, tmpFilePath string) error

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
