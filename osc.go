package main

import "github.com/hypebeast/go-osc/osc"

func OSCClient(config Config) (*osc.Client){
	return osc.NewClient(config.Resolume_IP, int(config.Resolume_Port)) 
}