package storage

import (
	"os"
	"testing"

	"github.com/pavlov061356/http_based_file_storage/internal/helpers"
)

const testFileSize = 1 << 20

func BenchmarkStorageRead(b *testing.B) {

	storage, err := NewStorage("/tmp")
	if err != nil {
		b.Fatal(err)
	}

	fileContent := make([]byte, testFileSize)

	for i := 0; i < testFileSize; i++ {
		fileContent[i] = byte(i)
	}

	previousFile := helpers.GetFilePath("/tmp", "test")
	_, err = os.Stat(previousFile)

	if err != nil && !os.IsNotExist(err) {
		b.Fatal(err)
		err = os.Remove(previousFile)

		if err != nil {
			b.Fatal(err)
		}
	}

	err = storage.saveFile("test", fileContent)

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {
			// defer wg.Done()
			tempFile, err := storage.Read("test")
			if err != nil {
				b.Fatal(err)
			}

			b.StopTimer()

			err = os.Remove(tempFile)

			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()
		}
	})
}

func BenchmarkStorageWrite(b *testing.B) {
	storage, err := NewStorage("/tmp")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {
			err = storage.saveFile("test", []byte("test"))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
