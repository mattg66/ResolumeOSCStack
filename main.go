package main

import (
	"fmt"
	"log"

	"github.com/hypebeast/go-osc/osc"
)

func main() {
	config := LoadConfig()

	client, err := ResolumeWS(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
    
	go func() {
		for msg := range client.Messages() {
			log.Printf("Got: %s", msg)
		}
	}()

	// sendError := c.WriteMessage(websocket.TextMessage, []byte("test"))
	// if sendError != nil {
	//     log.Println("Resolume Connection Error:", sendError)
	// }
	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/column", func(msg *osc.Message) {
		osc.PrintMessage(msg)
	})

	oscAddr := "0.0.0.0:" + fmt.Sprintf("%d", config.OSC_Port)
	server := &osc.Server{
		Addr:       oscAddr,
		Dispatcher: d,
	}
	log.Printf("Starting OSC server on %s", oscAddr)
	server.ListenAndServe()
}
