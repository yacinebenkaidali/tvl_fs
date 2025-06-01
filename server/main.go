package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	cmManager "github.com/yacinebenkaidali/tlv_tcp_server/cmmanager"
)

func main() {

	cmConfig := cmManager.ConnectionMangerConfig{
		OnConnect: func(conn *net.Conn) {
		},
		OnMessage: func(conn *net.Conn, data []byte) {

		},
	}
	cm := cmManager.NewConnectionManager(&cmConfig)
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM)

	cm.StartServer(":8000")

	<-quitCh
	// cm.Shutdown()
	log.Println("Received kill signal, closing off connection and server")
}
