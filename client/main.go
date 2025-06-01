package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

const (
	fileName = "./testdata/bigfile.dat"

	UPLOAD_CMD   uint16 = 0x0001
	DELETE_CMD   uint16 = 0x0002
	ARCHIVE_CMD  uint16 = 0x0003
	COMPRESS_CMD uint16 = 0x0004
	READ_CMD     uint16 = 0x0005
)

func main() {
	flag.Parse()
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		log.Fatalf("Failed to get file size: %v", err)
	}
	// Write command
	buffInfo := make([]byte, 8)
	binary.BigEndian.PutUint16(buffInfo[:2], UPLOAD_CMD)
	conn.Write(buffInfo[:2])

	// Write filename length
	binary.BigEndian.PutUint32(buffInfo[:4], uint32(len(fileName)))
	conn.Write(buffInfo[:4])
	// Write filename
	conn.Write([]byte(fileName))

	// Write fileSize
	binary.BigEndian.PutUint64(buffInfo, uint64(fi.Size()))
	_, err = conn.Write(buffInfo)
	if err != nil {
		log.Fatalf("Error writing file size: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buff := make([]byte, 4*1024)
		for {
			n, err := conn.Read(buff)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				log.Printf("error reading from connection %v\r\n", err)
				break
			}
			if n > 0 {
				log.Println(string(buff[:n]))
			}
		}
	}()

	// Write file
	reader := bufio.NewReader(f)
	buff := make([]byte, 4096)
	for {
		n, err := reader.Read(buff)
		if err != nil {
			if io.EOF == err || io.ErrUnexpectedEOF == err {
				break
			}
		}
		if n > 0 {
			_, err = conn.Write(buff[:n]) // Note: using buff[:n] to write only what was read
			if err != nil {
				switch err := err.(type) {
				case *net.OpError:
					// Handles network operation errors like connection refused, timeout
					log.Fatalf("Network operation error: %v", err)
					return
				default:
					// Handle other errors like:
					// - Connection reset by peer
					// - Broken pipe
					// - Connection timeout
					log.Fatalf("Write error: %v", err)
				}
			}
		}
	}
	wg.Wait()
	log.Println("Done transfering file")
}
