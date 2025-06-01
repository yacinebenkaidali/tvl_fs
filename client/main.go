package main

import (
	"flag"
	"log"
	"time"

	cmManager "github.com/yacinebenkaidali/tlv_tcp_client/cmmanager"
)

const (
	fileName = "./testdata/bigfile.dat"
)

func main() {
	flag.Parse()

	config := cmManager.ClientConfig{
		ReadTimeout:  25 * time.Second,
		WriteTimeout: 10 * time.Second,
		OnMessage: func(data []byte) {
			log.Printf("received %s\r\n", string(data))
		},
	}
	client := cmManager.NewConnectionClient(&config)
	client.Connect("localhost:8000")

	client.Wg.Add(1)
	go client.Listen()

	client.Wg.Add(1)
	go client.Upload(fileName)

	client.Wg.Wait()
}
