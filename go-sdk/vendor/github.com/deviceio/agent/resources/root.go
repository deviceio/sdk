package resources

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/deviceio/hmapi"
	"github.com/spf13/viper"
)

type root struct {
}

func (t *root) get(rw http.ResponseWriter, r *http.Request) {
	parentPath := r.Header.Get("X-Deviceio-Parent-Path")

	resource := &hmapi.Resource{
		Links: map[string]*hmapi.Link{
			"process": &hmapi.Link{
				Type: hmapi.MediaTypeJSON,
				Href: parentPath + "/process",
			},
			"filesystem": &hmapi.Link{
				Type: hmapi.MediaTypeJSON,
				Href: parentPath + "/filesystem",
			},
		},
		Content: map[string]*hmapi.Content{
			"id": &hmapi.Content{
				Type:  hmapi.MediaTypeHMAPIString,
				Value: viper.GetString("id"), //TODO: this needs to be passed in
			},
			"hostname": &hmapi.Content{
				Type: hmapi.MediaTypeHMAPIString,
				Value: (func() string {
					hostname, _ := os.Hostname()
					return hostname
				})(),
			},
			"architecture": &hmapi.Content{
				Type:  hmapi.MediaTypeHMAPIString,
				Value: runtime.GOARCH,
			},
			"platform": &hmapi.Content{
				Type:  hmapi.MediaTypeHMAPIString,
				Value: runtime.GOOS,
			},
		},
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(&resource)
}

func (t *root) chunkedResponseTest(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)

	httpflush := rw.(http.Flusher).Flush
	httpclose := rw.(http.CloseNotifier).CloseNotify()
	pipeReader, pipeWriter := io.Pipe()

	httpflush()

	go func() {
		for {
			select {
			case <-httpclose:
				pipeWriter.Close()
				return
			default:
				pipeWriter.Write([]byte("test"))
				httpflush()
			}

			time.Sleep(1 * time.Second)
		}
	}()

	io.Copy(rw, pipeReader)
}
