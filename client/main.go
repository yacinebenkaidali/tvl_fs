package main

import (
	"flag"
	"log"
	"time"

	cmManager "github.com/yacinebenkaidali/tlv_tcp_client/cmmanager"
)

func main() {
	var cmd cmManager.Command = cmManager.READ_CMD

	filename := flag.String("f", "./testdata/bigfile.dat", "the name of the file to send")
	flag.Var(&cmd, "cmd", "Command to execute (upload, delete, archive, compress, read)")

	flag.Parse()

	config := cmManager.ClientConfig{
		ReadTimeout:  25 * time.Second,
		WriteTimeout: 10 * time.Second,
		OnMessage: func(cmd cmManager.Command, data []byte) {
			switch cmd {
			case cmManager.READ_CMD:
				{
					// f, err := os.Create(fmt.Sprintf("./received/%s", *filename))
					// // rebuild file
				}
			default:
				{
					log.Printf("received %s\r\n", string(data))
				}
			}
		},
	}
	client := cmManager.NewConnectionClient(&config)
	client.Connect("localhost:8000")

	client.Wg.Add(1)
	go client.Listen(cmd, *filename)

	client.Wg.Add(1)
	switch cmd {
	case cmManager.READ_CMD:
		go client.Read(*filename)
	case cmManager.UPLOAD_CMD:
		go client.Upload(*filename)
	}

	client.Wg.Wait()

	client.Shutdown()
}
