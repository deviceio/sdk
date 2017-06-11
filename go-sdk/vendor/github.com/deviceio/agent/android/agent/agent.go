package agent

import (
	"log"

	_ "github.com/deviceio/agent/resources/filesystem"
	"github.com/deviceio/agent/transport"
	"github.com/deviceio/shared/logging"
)

func Start(agentid string, host string, port int, insecure bool) {
	log.Println("Starting Agent")

	transport.NewConnection(&logging.DefaultLogger{}).Dial(&transport.ConnectionOpts{
		ID:   agentid,
		Tags: []string{},
		TransportAllowSelfSigned: insecure,
		TransportHost:            host,
		TransportPort:            port,
	})

	log.Println("Transport Started")

	<-make(chan bool)
}
