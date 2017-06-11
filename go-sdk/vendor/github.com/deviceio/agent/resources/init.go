package resources

import "github.com/deviceio/agent/transport"

func init() {
	r := &root{}
	transport.Router.HandleFunc("/", r.get).Methods("GET")
	transport.Router.HandleFunc("/chunk-response-test", r.chunkedResponseTest).Methods("GET")
}
