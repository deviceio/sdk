package process

import (
	"sync"

	"github.com/deviceio/agent/transport"
)

func init() {
	root := &Root{
		Itemsmu: &sync.Mutex{},
		Items:   map[string]*process{},
	}

	transport.Router.HandleFunc("/process", root.Get).Methods("GET")
	transport.Router.HandleFunc("/process", root.CreateProcess).Methods("POST")
	transport.Router.HandleFunc("/process/{proc-id}", root.GetProcess).Methods("GET")
	transport.Router.HandleFunc("/process/{proc-id}/start", root.StartProcess).Methods("POST")
	transport.Router.HandleFunc("/process/{proc-id}/stop", root.StopProcess).Methods("POST")
	transport.Router.HandleFunc("/process/{proc-id}/stdin", root.StdinProcess).Methods("POST")
	transport.Router.HandleFunc("/process/{proc-id}/stdout", root.StdoutProcess).Methods("GET")
	transport.Router.HandleFunc("/process/{proc-id}/stderr", root.StderrProcess).Methods("GET")
	transport.Router.HandleFunc("/process/{proc-id}", root.DeleteProcess).Methods("DELETE")
}
