package filesystem

import (
	"github.com/deviceio/agent/transport"
	"github.com/deviceio/shared/logging"
)

func init() {
	fs := &Root{
		logger: &logging.DefaultLogger{},
	}

	transport.Router.HandleFunc("/filesystem", fs.Get).Methods("GET")
	transport.Router.HandleFunc("/filesystem/read", fs.Read).Methods("POST")
	transport.Router.HandleFunc("/filesystem/write", fs.Write).Methods("POST")
}
