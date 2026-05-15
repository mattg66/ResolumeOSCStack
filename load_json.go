package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"os"
)

type Config struct {
	Resolume_IP   string `json:"resolume_ip" validate:"required,ip4_addr"`
	Resolume_Port uint   `json:"resolume_port" validate:"required,port"`
	OSC_Port      uint   `json:"osc_port" validate:"required,port"`
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
