package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hypebeast/go-osc/osc"
)

type CompositionConfig struct {
	Name struct {
		Value string `json:"value"`
	} `json:"name"`
	Layers []struct {
		ID         int64 `json:"id"`
		Transition struct {
			Duration struct {
				ID int `json:"id"`
			} `json:"duration"`
		} `json:"transition"`
	} `json:"layers"`
}
type safeClient struct {
	mu sync.RWMutex
	c  *WSClient
}
type safeConfig struct {
	mu            sync.RWMutex
	runningConfig *CompositionConfig
}

func (s *safeConfig) set(c *CompositionConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runningConfig = c
}
func (s *safeConfig) get() *CompositionConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runningConfig
}
func (s *safeClient) Send(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.c.Send(data)
}

func (s *safeClient) set(c *WSClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.c = c
}

func (s *safeClient) get() *WSClient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c
}

type WSAction struct {
	Action    string      `json:"action"`
	Parameter string      `json:"parameter,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}

func main() {
	config := LoadConfig()
	runningConfig := &safeConfig{}
	sc := &safeClient{}

	connect(config, sc, runningConfig)

	d := osc.NewStandardDispatcher()

	d.AddMsgHandler("*", func(msg *osc.Message) {

		if strings.HasPrefix(msg.Address, "/column/") {
			if runningConfig.get() == nil {
				log.Printf("No composition config received yet, ignoring /column message")
				return
			}
			if len(msg.Arguments) < 1 {
				log.Printf("Invalid /column message: expected at least 1 argument, got %d", len(msg.Arguments))
				return
			}

			parts := strings.Split(msg.Address, "/")
			if len(parts) < 3 {
				log.Printf("Invalid /column path: %s", msg.Address)
				return
			}

			var columnNum int32
			_, err := fmt.Sscanf(parts[2], "%d", &columnNum)
			if err != nil {
				log.Printf("Invalid column ID in path: %s", parts[2])
				return
			}

			var value float32
			switch v := msg.Arguments[0].(type) {
			case float32:
				value = v
			case int32:
				value = float32(v)
			default:
				log.Printf("Invalid /column message: argument is not a float32 or int32")
				return
			}

			log.Printf("Column %d: Transition: %.2vs", columnNum, value)

			for index := range runningConfig.get().Layers {
				layer := WSAction{
					Action:    "set",
					Parameter: fmt.Sprintf("/parameter/by-id/%d", runningConfig.get().Layers[index].Transition.Duration.ID),
					Value:     value,
				}
				layerData, err := json.Marshal(layer)
				if err != nil {
					log.Printf("Error marshaling JSON: %v", err)
					return
				}
				sc.Send(layerData)
			}

			action := WSAction{
				Action:    "trigger",
				Parameter: fmt.Sprintf("/composition/columns/%d/connect", columnNum),
			}

			jsonData, err := json.Marshal(action)
			if err != nil {
				log.Printf("Error marshaling JSON: %v", err)
				return
			}
			sc.Send(jsonData)
			return
		} else if config.QLab != nil {
			qlabClient := osc.NewClient(config.QLab.IP, int(config.QLab.OSCPort))
			log.Printf("Sent to QLab: %v", msg)
			qlabClient.Send(msg)
		}
	})
	oscAddr := "0.0.0.0:" + fmt.Sprintf("%d", config.OSCListenPort)
	server := &osc.Server{
		Addr:       oscAddr,
		Dispatcher: d,
	}
	log.Printf("Starting OSC server on %s", oscAddr)
	server.ListenAndServe()
}
