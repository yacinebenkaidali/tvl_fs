package cmmanager

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"time"
)

const (
	UPLOAD_CMD   uint16 = 0x0001
	DELETE_CMD   uint16 = 0x0002
	ARCHIVE_CMD  uint16 = 0x0003
	COMPRESS_CMD uint16 = 0x0004
	READ_CMD     uint16 = 0x0005
)

type Connection struct {
	ID         string
	conn       net.Conn
	Address    string
	progressCh chan float32
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type ConnectionManger struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	address     string
	listener    net.Listener
	connections sync.Map
	wg          sync.WaitGroup

	readTimeout  time.Duration
	writeTimeout time.Duration

	onMessage    func(conn *net.Conn, data []byte)
	onConnect    func(conn *net.Conn)
	onDisconnect func(conn *net.Conn)
}

type ConnectionMangerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	OnMessage    func(conn *net.Conn, data []byte)
	OnConnect    func(conn *net.Conn)
	OnDisconnect func(conn *net.Conn)
}

func NewConnectionManager(config *ConnectionMangerConfig) *ConnectionManger {
	ctx, cancel := context.WithCancel(context.Background())

	cm := &ConnectionManger{
		readTimeout:  config.ReadTimeout,
		writeTimeout: config.WriteTimeout,
		onConnect:    config.OnConnect,
		onDisconnect: config.OnDisconnect,
		onMessage:    config.OnMessage,
		ctx:          ctx,
		cancelFunc:   cancel,
	}

	if config.ReadTimeout == 0 {
		cm.readTimeout = 30 * time.Second
	}

	if config.WriteTimeout == 0 {
		cm.writeTimeout = 60 * time.Second
	}

	if config.OnDisconnect != nil {
		cm.onDisconnect = config.OnDisconnect
	}

	if config.OnConnect != nil {
		cm.onConnect = config.OnConnect
	}

	if config.OnMessage != nil {
		cm.onMessage = config.OnMessage
	}

	return cm
}

func (cm *ConnectionManger) StartServer(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	cm.address = address
	cm.listener = listener
	log.Printf("server started listening at %s\n", address)
	cm.wg.Add(1)

	go cm.acceptConnections()
	return nil
}

func (cm *ConnectionManger) acceptConnections() {
	defer cm.wg.Done()

	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
			{
				conn, err := cm.listener.Accept()
				if err != nil {
					log.Printf("there was a problem establishing a connection %+v\n", err)
					continue
				}
				log.Printf("connection established with client %s\n", conn.RemoteAddr())
				connectionId := fmt.Sprintf("incoming_conn_%s_%d", conn.RemoteAddr(), time.Now().UnixNano())
				cm.handleConnection(conn, connectionId)
			}
		}

	}
}

func (cm *ConnectionManger) handleConnection(conn net.Conn, id string) {
	ctx, cancel := context.WithCancel(cm.ctx)
	connection := Connection{
		ID:         id,
		conn:       conn,
		Address:    conn.RemoteAddr().String(),
		ctx:        ctx,
		cancelFunc: cancel,
		progressCh: make(chan float32, 10),
	}
	cm.connections.Store(id, connection)

	if cm.onConnect != nil {
		cm.onConnect(&conn)
	}

	cm.wg.Add(2)
	go cm.handleConnectionWrite(&connection)
	go cm.handleConnectionRead(&connection)
}

func (cm *ConnectionManger) handleConnectionWrite(connection *Connection) {
	defer cm.wg.Done()

	for percentage := range connection.progressCh {
		buff := make([]byte, 8)
		binary.BigEndian.PutUint64(buff, uint64(percentage))

		_, err := connection.conn.Write(buff)
		if err != nil {
			if errors.Is(err, syscall.EPIPE) {
				return
			}
			log.Printf("there was a problem writing to the connection, %q\n", err)
		}
		time.Sleep(time.Microsecond * 100)
	}

}

func (cm *ConnectionManger) handleConnectionRead(conn *Connection) {
	defer func() {
		cm.wg.Done()
		cm.closeConnection(conn)
	}()

	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
			{
				// READING_TYPE
				cmd, err := handleTypeParsing(conn.conn)
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
						log.Printf("Client disconnected: %s", (conn.conn).RemoteAddr())
						return
					}
					log.Printf("Error reading length prefix from %s: %v", (conn.conn).RemoteAddr(), err)
					return
				}
				// validation
				switch cmd {
				case UPLOAD_CMD, DELETE_CMD, READ_CMD, ARCHIVE_CMD, COMPRESS_CMD:
					log.Printf("received %d command from client %s", cmd, conn.conn.RemoteAddr().String())
					// valid command, continue processing
				default:
					log.Printf("unknown command received %d", cmd)
					return
				}

				fileNameLgth, err := handleLengthParsing(conn.conn, 4)
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
						log.Printf("Client disconnected: %s", (conn.conn).RemoteAddr())
						return
					}
					log.Printf("Error reading length prefix from %s: %v", (conn.conn).RemoteAddr(), err)
					return
				}
				fileName, err := handleFileNameParsing(conn.conn, int(fileNameLgth))
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
						log.Printf("Client disconnected: %s", (conn.conn).RemoteAddr())
						return
					}
					log.Printf("Error reading length prefix from %s: %v", (conn.conn).RemoteAddr(), err)
					return
				}
				var opError error
				switch cmd {
				case UPLOAD_CMD:
					{
						opError = uploadFile(fileName, conn)
					}
				case DELETE_CMD:
					opError = deleteFile(fileName)
				case READ_CMD:
					{
						opError = readFile(fileName, conn.conn)
						if opError == nil {
							conn.conn.Write(fmt.Appendf(nil, "Done streaming file %s to client %s\r\n", fileName, conn.conn.RemoteAddr()))
							cm.closeConnection(conn)
						}
					}
				case ARCHIVE_CMD:
					opError = archiveFile(fileName)
				case COMPRESS_CMD:
					opError = compressFile(fileName)
				}

				if opError != nil {
					if errors.Is(opError, io.EOF) || errors.Is(opError, io.ErrUnexpectedEOF) {
						log.Printf("Client disconnected: %s", (conn.conn).RemoteAddr())
						return
					}
					log.Printf("Error reading length prefix from %s: %v", (conn.conn).RemoteAddr(), opError)
					return
				}
			}
		}
	}
}

func (cm *ConnectionManger) closeConnection(conn *Connection) {
	if _, loaded := cm.connections.LoadAndDelete(conn.ID); !loaded {
		log.Printf("connection with id %s was not found\n", conn.ID)
	}

	if err := conn.conn.Close(); err != nil {
		log.Printf("there was an error closing connection with id %s\n", conn.ID)
	}

	conn.cancelFunc()
	if cm.onDisconnect != nil {
		cm.onDisconnect(&conn.conn)
	}
	close(conn.progressCh)
	log.Printf("connection with id %s was closed\n", conn.ID)
}
