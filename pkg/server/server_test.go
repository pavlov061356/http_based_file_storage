package server

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pavlov061356/http_based_file_storage/internal/helpers"
	"github.com/pavlov061356/http_based_file_storage/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestSaveFile(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write([]byte("test"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, _ := http.NewRequest("POST", "/file", b)
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	// req.Write(b)

	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
}

func TestSaveAlreadySavedFile(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write([]byte("test"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, _ := http.NewRequest("POST", "/file", b)
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	// req.Write(b)

	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
}

func TestSaveFileWithEmptyFile(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write(nil)

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, _ := http.NewRequest("POST", "/file", b)
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	// req.Write(b)

	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
}

func TestGetFile(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write([]byte("test"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, _ := http.NewRequest("POST", "/file", b)
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	// req.Write(b)

	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/file/"+response["hash"].(string), nil)
	r.ServeHTTP(w, req)
	fmt.Println(w.Body.String())
	assert.Equal(t, 200, w.Code)
}

func TestGetFileONonExistentFile(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/file/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestDeleteFile(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write([]byte("test"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, err := http.NewRequest("POST", "/file", b)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	w = httptest.NewRecorder()

	req, err = http.NewRequest("DELETE", "/file/"+response["hash"].(string), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()

	req, err = http.NewRequest("DELETE", "/file/"+response["hash"].(string), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestDeleteWithEmptyHash(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/file/", nil)

	if err != nil {
		t.Fatal(err)
	}

	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestDeleteWithInvalidHash(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "/file/invalid_hash", nil)

	if err != nil {
		t.Fatal(err)
	}

	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestPostFileWithHashHeaders(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write([]byte("test"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, _ := http.NewRequest("POST", "/file", b)
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Add("SHA256", helpers.GetFileHash(sha256.New(), file))
	req.Header.Add("SHA1", helpers.GetFileHash(sha1.New(), file))
	req.Header.Add("SHA512", helpers.GetFileHash(sha512.New(), file))
	req.Header.Add("MD5", helpers.GetFileHash(md5.New(), file))

	// req.Write(b)

	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)
}

func TestPostFileWithWrongHashHeaders(t *testing.T) {
	os.RemoveAll("/tmp/store")
	storage, err := storage.NewStorage("/tmp")

	if err != nil {
		t.Fatal(err)
	}

	server, err := NewHTTPFileStorageServer(
		storage,
		&Config{
			Host:        "localhost",
			Port:        8080,
			StoragePath: "/tmp",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := server.setupRouter()

	w := httptest.NewRecorder()

	b := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(b)

	part, err := multipartWriter.CreateFormFile("file", "test")

	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp("", "test")

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Write([]byte("test"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}
	multipartWriter.Close()

	req, _ := http.NewRequest("POST", "/file", b)
	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Add("SHA256", helpers.GetFileHash(sha256.New(), file))
	req.Header.Add("SHA1", helpers.GetFileHash(sha1.New(), file))
	req.Header.Add("SHA512", helpers.GetFileHash(sha512.New(), file))
	req.Header.Add("MD5", "asdfghjk")

	// req.Write(b)

	r.ServeHTTP(w, req)
	assert.Equal(t, 412, w.Code)
}
