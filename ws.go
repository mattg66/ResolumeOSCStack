package main

import "github.com/gorilla/websocket"
import "net/url"
import "log"
import "fmt"
import "encoding/json"

type WSClient struct {
	conn     *websocket.Conn
	incoming chan []byte
	outgoing chan []byte
	done     chan struct{}
}

func ResolumeWS(config Config) (*WSClient, error) {
	resolumeAddress := config.Resolume.IP + ":" + fmt.Sprintf("%d", config.Resolume.WebsocketPort)
	u := url.URL{Scheme: "ws", Host: resolumeAddress, Path: "/api/v1"}
	log.Printf("Connecting to Resolume Arena %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	} else {
		log.Printf("Connected to Resolume Arena %s", u.String())
	}
	client := &WSClient{
		conn:     c,
		incoming: make(chan []byte, 10),
		outgoing: make(chan []byte, 10),
		done:     make(chan struct{}),
	}

	go client.readPump()
	go client.writePump()

	return client, nil
}

func (ws *WSClient) readPump() {
	defer close(ws.done)
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			return
		}
		ws.incoming <- message
	}
}

func (ws *WSClient) writePump() {
	for msg := range ws.outgoing {
		err := ws.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func (ws *WSClient) Send(msg []byte) {
	ws.outgoing <- msg
}

func (ws *WSClient) Messages() <-chan []byte {
	return ws.incoming
}

func (ws *WSClient) Close() {
	close(ws.outgoing)
	ws.conn.Close()
}

func connect(config Config, sc *safeClient, runningConfig *safeConfig) {
	client, err := ResolumeWS(config)
	if err != nil {
		log.Fatal(err)
	}
	sc.set(client)

	sc.get().conn.SetCloseHandler(func(code int, text string) error {
		log.Println("Connection closed to Resolume closed")
		connect(config, sc, runningConfig)

		return nil
	})
	go func() {
		for msg := range client.Messages() {
			var receivedConfig CompositionConfig
			if err := json.Unmarshal(msg, &receivedConfig); err != nil || receivedConfig.Name.Value == "" {
				continue
			}
			existing := runningConfig.get()
			if existing == nil || existing.Name != receivedConfig.Name || len(existing.Layers) != len(receivedConfig.Layers) {
				runningConfig.set(&receivedConfig)
				log.Printf("Composition: %s (%d layers)", receivedConfig.Name.Value, len(receivedConfig.Layers))
			}
		}
	}()
}
