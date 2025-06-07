package cmmanager

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	UPLOAD_CMD   Command = 0x0001
	DELETE_CMD   Command = 0x0002
	ARCHIVE_CMD  Command = 0x0003
	COMPRESS_CMD Command = 0x0004
	READ_CMD     Command = 0x0005
)

type ConnectionClient struct {
	address string
	conn    *net.Conn
	Wg      sync.WaitGroup
	ctx     context.Context
	Cancel  context.CancelFunc

	readTimeout  time.Duration
	writeTimeout time.Duration

	onMessage func(cmd Command, data []byte)
}

type ClientConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	OnMessage func(cmd Command, data []byte)
}

func NewConnectionClient(config *ClientConfig) *ConnectionClient {
	ctx, cancel := context.WithCancel(context.Background())
	cm := ConnectionClient{
		ctx:          ctx,
		Cancel:       cancel,
		Wg:           sync.WaitGroup{},
		readTimeout:  config.ReadTimeout,
		writeTimeout: config.WriteTimeout,
		onMessage:    config.OnMessage,
	}

	if config.ReadTimeout == 0 {
		cm.readTimeout = 30 * time.Second
	}

	if config.WriteTimeout == 0 {
		cm.writeTimeout = 60 * time.Second
	}

	return &cm
}

func (c *ConnectionClient) Connect(address string) error {
	c.address = address
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.conn = &conn

	return nil
}

func (c *ConnectionClient) Listen(cmd Command, filename string) {
	defer c.Wg.Done()

	deadline := time.Now().Add(c.readTimeout)
	err := (*c.conn).SetReadDeadline(deadline)
	if err != nil {
		log.Printf("error set from connection %v\r\n", err)
		return
	}

	f, err := os.Create(fmt.Sprintf("./received/%s", filepath.Base(filename)))
	if err != nil {
		log.Printf("there was a problem opening a file %+v\r\n", err)
	}
	defer f.Close()

	buff := make([]byte, 4*1024)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			{
				n, err := (*c.conn).Read(buff)
				if err != nil {
					if io.EOF == err {
						log.Printf("connection with server closed\r\n")
						return
					}
					log.Printf("error reading from connection %v\r\n", err)
				}
				if n > 0 {
					if cmd == READ_CMD {
						_, err = f.Write(buff[:n])
						if err != nil {
							log.Printf("there was a problme writing to file %+v\r\n", err)
						}
					} else {
						if c.onMessage != nil {
							c.onMessage(cmd, buff[:n])
						}
					}
				}
			}
		}
	}
}

func (c *ConnectionClient) Upload(filename string) error {
	defer func() {
		c.Wg.Done()
		log.Println("returning ....")
	}()
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	// Write command
	buffInfo := make([]byte, 8)
	binary.BigEndian.PutUint16(buffInfo[:2], uint16(UPLOAD_CMD))
	(*c.conn).Write(buffInfo[:2])

	// Write filename length
	binary.BigEndian.PutUint32(buffInfo[:4], uint32(len(filename)))
	(*c.conn).Write(buffInfo[:4])

	// Write filename
	(*c.conn).Write([]byte(filename))

	// Write fileSize
	binary.BigEndian.PutUint64(buffInfo, uint64(fi.Size()))
	_, err = (*c.conn).Write(buffInfo)
	if err != nil {
		return err
	}

	// Write file content
	buff := make([]byte, 4096)
	for {
		n, err := f.Read(buff)
		if err != nil {
			return err
		}
		if n > 0 {
			deadline := time.Now().Add(c.writeTimeout)
			err = (*c.conn).SetWriteDeadline(deadline)
			if err != nil {
				return err
			}
			_, err = (*c.conn).Write(buff[:n]) // Note: using buff[:n] to write only what was read
			if err != nil {
				return err
			}
		}
	}
}

func (c *ConnectionClient) Read(filename string) error {
	defer c.Wg.Done()
	// Write command
	buffInfo := make([]byte, 8)
	binary.BigEndian.PutUint16(buffInfo[:2], uint16(READ_CMD))
	_, err := (*c.conn).Write(buffInfo[:2])
	if err != nil {
		return err
	}
	// Write filename length
	binary.BigEndian.PutUint32(buffInfo[:4], uint32(len(filename)))
	_, err = (*c.conn).Write(buffInfo[:4])
	if err != nil {
		return err
	}

	// Write filename
	_, err = (*c.conn).Write([]byte(filename))
	return err
}

func (c *ConnectionClient) Shutdown() {
	if err := (*c.conn).Close(); err != nil {
		log.Printf("there was a problem closing this connection\r\n")
	}
	c.Cancel()
}
