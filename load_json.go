package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"os"
)

type QLab struct {
	IP      string `json:"ip" validate:"required,ip4_addr"`
	OSCPort uint   `json:"osc_port" validate:"required,port"`
}

type Resolume struct {
	IP            string `json:"ip" validate:"required,ip4_addr"`
	WebsocketPort uint   `json:"websocket_port" validate:"required,port"`
}

type Config struct {
	OSCListenPort uint     `json:"osc_listen_port" validate:"required,port"`
	Resolume      Resolume `json:"resolume" validate:"required"`
	QLab          *QLab   `json:"qlab,omitempty" validate:"omitempty"`
}

func LoadJSON[T any](filename string) (T, error) {
	var data T
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(fileData, &data)
	return data, err
}

func LoadConfig() Config {
	validate := validator.New()
	data, err := LoadJSON[Config]("config.json")
	if err != nil {
		fmt.Print("Invalid config.json\n")
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	}
	err = validate.Struct(data)

	if err != nil {
		fmt.Println(err.(validator.ValidationErrors))
		fmt.Print("Invalid config.json\n")
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	}

	return data
}
