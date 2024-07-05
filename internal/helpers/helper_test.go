package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilePath(t *testing.T) {
	filePath := GetFilePath("/tmp", "hash")
	assert.Equal(t, "/tmp/store/ha/hash", filePath)
}

func TestGetFileParentPath(t *testing.T) {
	fileParentPath := GetFileParentPath("/tmp", "hash")
	assert.Equal(t, "/tmp/store/ha", fileParentPath)
}
