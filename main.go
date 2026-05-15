package main

import (
	"encoding/json"
	"fmt"
	"log"
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

type WSAction struct {
	Action    string      `json:"action"`
	Parameter string      `json:"parameter,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}

func main() {
	config := LoadConfig()
	var runningConfig CompositionConfig
	client, err := ResolumeWS(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	go func() {
		for msg := range client.Messages() {
			var receivedConfig CompositionConfig
			err = json.Unmarshal(msg, &receivedConfig)
			if err == nil && receivedConfig.Name.Value != "" {
				if (runningConfig.Name != receivedConfig.Name) || (len(runningConfig.Layers) != len(receivedConfig.Layers)) {
					runningConfig = receivedConfig
					log.Printf("Composition: %s (%d layers)", runningConfig.Name.Value, len(runningConfig.Layers))
				}
			}
		}
	}()

	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/column", func(msg *osc.Message) {
		if len(msg.Arguments) < 2 {
			log.Printf("Invalid /column message: expected 2 arguments, got %d", len(msg.Arguments))
			return
		}

		columnNum, ok := msg.Arguments[0].(int32)
		if !ok {
			log.Printf("Invalid /column message: first argument is not an int32")
			return
		}

		var value float32
		switch v := msg.Arguments[1].(type) {
		case float32:
			value = v
		case int32:
			value = float32(v)
		default:
			log.Printf("Invalid /column message: second argument is not a float32 or int32")
			return
		}

		log.Printf("Column %d: Transition: %.2vs", columnNum, value)

		action := WSAction{
			Action:    "trigger",
			Parameter: fmt.Sprintf("/composition/columns/%d/connect", columnNum),
		}

		jsonData, err := json.Marshal(action)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			return
		}
		for index := range runningConfig.Layers {
			layer := WSAction{
				Action:    "set",
				Parameter: fmt.Sprintf("/parameter/by-id/%d", runningConfig.Layers[index].Transition.Duration.ID),
				Value:     value,
			}
			layerData, err := json.Marshal(layer)
			if err != nil {
				log.Printf("Error marshaling JSON: %v", err)
				return
			}
			client.Send(layerData)
		}

		client.Send(jsonData)
	})

	oscAddr := "0.0.0.0:" + fmt.Sprintf("%d", config.OSC_Port)
	server := &osc.Server{
		Addr:       oscAddr,
		Dispatcher: d,
	}
	log.Printf("Starting OSC server on %s", oscAddr)
	server.ListenAndServe()
}
