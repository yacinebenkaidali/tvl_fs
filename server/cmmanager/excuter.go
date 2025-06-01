package cmmanager

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

func archiveFile(fileName string) error {
	return nil
}

func deleteFile(fileName string) error {
	return nil
}

func compressFile(fileName string) error {
	return nil
}

func readFile(fileName string, w io.Writer) error {
	return nil
}

func uploadFile(fileName string, r *net.Conn) error {
	const (
		defaultBufferSize = 4 * 1024
		maxFileSize       = 1 * 1024 * 1024 * 1024 // 1GB max file size
	)
	lengthBytes := make([]byte, 8)
	_, err := io.ReadFull(*r, lengthBytes)
	if err != nil {
		return err
	}
	fileLength := binary.BigEndian.Uint64(lengthBytes)
	// TODO: think of a better folder structure, CAS ??
	f, err := os.Create(fmt.Sprintf("testdata/%s", filepath.Base(fileName)))
	if err != nil {
		return err
	}

	defer f.Close()
	var currentReadSize uint64 = 0

	buff := make([]byte, defaultBufferSize)
	for {
		var readSize = min(fileLength-currentReadSize, defaultBufferSize)
		n, err := io.ReadFull(*r, buff[:readSize])
		if err != nil {
			return err
		}
		if n > 0 {
			nn, err := f.Write(buff[:n])
			if err != nil {
				return err
			}
			currentReadSize += uint64(nn)
			if currentReadSize == fileLength {
				// end of the current file, the content should have been written to local file
				break
			}
		}
	}
	if err := f.Close(); err != nil {
		log.Println("There was a problem flushing content to disk")
	}
	if currentReadSize == fileLength {
		log.Printf("Successfully wrote file %s (%d bytes)", fileName, currentReadSize)
	} else {
		log.Printf("Incomplete file transfer %s: got %d of %d bytes", fileName, currentReadSize, fileLength)
		if err := os.Remove(fileName); err != nil {
			log.Printf("Error removing incomplete file %s: %v", fileName, err)
		}
	}
	return nil
}
