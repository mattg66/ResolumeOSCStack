package main

import "fmt"
import "github.com/hypebeast/go-osc/osc"

func main() {
	config := LoadConfig()
    addr := config.Resolume_IP + ":" + fmt.Sprintf("%d", config.Resolume_Port)
	resolumeClient := OSCClient(config)
	
    d := osc.NewStandardDispatcher()
    d.AddMsgHandler("/column", func(msg *osc.Message) {
        osc.PrintMessage(msg)
    })

    server := &osc.Server{
        Addr: addr,
        Dispatcher:d,
    }
    server.ListenAndServe()
}
