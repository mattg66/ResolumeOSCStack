package main

import "github.com/gorilla/websocket"
import "net/url"
import "log"
import "fmt"

type WSClient struct {
	conn *websocket.Conn
	incoming chan []byte
	outgoing chan []byte
	done	chan struct{}
}
func ResolumeWS(config Config) (*WSClient, error) {
	resolumeAddress := config.Resolume_IP + ":" + fmt.Sprintf("%d", config.Resolume_Port)
	u := url.URL{Scheme: "ws", Host: resolumeAddress, Path: "/api/v1"}
	log.Printf("Connecting to Resolume Arena %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &WSClient{
		conn:	c,
		incoming:	make(chan []byte, 10),
		outgoing:	make(chan []byte, 10),
		done: 	make(chan struct{}),
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
			log.Println("read:", err)
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