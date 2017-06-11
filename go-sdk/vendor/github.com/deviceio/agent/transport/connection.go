package transport

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/shared/logging"
	"github.com/hashicorp/yamux"
	"github.com/jpillora/backoff"
)

// Connection represents our upstream connection to a hub.
type Connection struct {
	opts      *ConnectionOpts
	reconnect int
	jitter    int
	backoff   *backoff.Backoff
	logger    logging.Logger
}

// NewConnection creates a new instance of the Connection type
func NewConnection(logger logging.Logger) *Connection {
	return &Connection{
		logger: logger,
	}
}

// Dial attempts to connect to the upstream hub. If dialing fails a backoff
// algorithm is applied during reconnection attempts to alleviate load on a hub
// that disappears momentarily.
func (t *Connection) Dial(opts *ConnectionOpts) {
	t.opts = opts

	t.backoff = &backoff.Backoff{
		Max:    5 * time.Second,
		Jitter: true,
	}

	for {
		err := t.run()

		logrus.WithFields(logrus.Fields{
			"host":  t.opts.TransportHost,
			"port":  t.opts.TransportPort,
			"error": err.Error(),
		}).Warn("transport failure")

		wait := t.backoff.Duration()

		if wait >= t.backoff.Max {
			t.backoff.Reset()
		}

		logrus.WithFields(logrus.Fields{
			"host": t.opts.TransportHost,
			"port": t.opts.TransportPort,
			"wait": wait,
		}).Info("transport retry")

		time.Sleep(wait)
	}
}

// run conducts the setup of the multiplexed tcp stream server to the hub and registers
// the base http server to be served over the multiplexed connection.
func (t *Connection) run() error {
	dialaddr := fmt.Sprintf("%v:%v", t.opts.TransportHost, t.opts.TransportPort)

	conn, err := tls.Dial("tcp", dialaddr, &tls.Config{
		InsecureSkipVerify: t.opts.TransportAllowSelfSigned,
	})

	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"localAddr":  conn.LocalAddr(),
		"remoteAddr": conn.RemoteAddr(),
	}).Info("transport up")

	server, err := yamux.Server(conn, nil)

	if err != nil {
		return err
	}

	Router.HandleFunc("/info", t.httpGetInfo)

	return http.Serve(server, Router)
}

// httpGetInfo provides basic information about this device a hub needs to properly
// manage the transport connection.
func (t *Connection) httpGetInfo(w http.ResponseWriter, r *http.Request) {
	type info struct {
		ID           string
		Hostname     string
		Architecture string
		Platform     string
		Tags         []string
	}

	hostname, err := os.Hostname()

	if err != nil {
		hostname = "Unknown"
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&info{
		ID:           t.opts.ID,
		Tags:         t.opts.Tags,
		Hostname:     hostname,
		Architecture: runtime.GOARCH,
		Platform:     runtime.GOOS,
	})

	if err != nil {
		t.logger.Error(err.Error())
		w.WriteHeader(500)
		w.Write([]byte(""))
		return
	}
}
