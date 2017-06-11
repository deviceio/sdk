package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"

	"bytes"

	"mime/multipart"

	"github.com/deviceio/hmapi"
)

type process struct {
	id            string
	cmd           *exec.Cmd
	mu            *sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	stdoutPipe    io.ReadCloser
	stderrPipe    io.ReadCloser
	stdinPipe     io.WriteCloser
	stdoutReaders []chan []byte
	stderrReaders []chan []byte
}

func (t *process) get(rw http.ResponseWriter, r *http.Request) {
	parentPath := r.Header.Get("X-Deviceio-Parent-Path")

	resource := &hmapi.Resource{
		Links: map[string]*hmapi.Link{
			"self": &hmapi.Link{
				Href: parentPath + "/process/" + t.id,
			},
			"parent": &hmapi.Link{
				Href: parentPath + "/process",
			},
			"stdout": &hmapi.Link{
				Href: fmt.Sprintf("%v/process/%v/stdout", parentPath, t.id),
				Type: hmapi.MediaTypeOctetStream,
			},
			"stderr": &hmapi.Link{
				Href: fmt.Sprintf("%v/process/%v/stderr", parentPath, t.id),
				Type: hmapi.MediaTypeOctetStream,
			},
		},
		Forms: map[string]*hmapi.Form{
			"delete": &hmapi.Form{
				Action:  fmt.Sprintf("%v/process/%v", parentPath, t.id),
				Enctype: hmapi.MediaTypeMultipartFormData,
				Method:  hmapi.DELETE,
			},
			"start": &hmapi.Form{
				Action:  fmt.Sprintf("%v/process/%v/start", parentPath, t.id),
				Enctype: hmapi.MediaTypeMultipartFormData,
				Method:  hmapi.POST,
			},
			"stop": &hmapi.Form{
				Action:  fmt.Sprintf("%v/process/%v/stop", parentPath, t.id),
				Enctype: hmapi.MediaTypeMultipartFormData,
				Method:  hmapi.POST,
			},
			"stdin": &hmapi.Form{
				Action:  fmt.Sprintf("%v/process/%v/stdin", parentPath, t.id),
				Method:  hmapi.POST,
				Enctype: hmapi.MediaTypeMultipartFormData,
				Fields: []*hmapi.FormField{
					&hmapi.FormField{
						Name:     "data",
						Type:     hmapi.MediaTypeOctetStream,
						Required: true,
					},
				},
			},
		},
		Content: map[string]*hmapi.Content{
			"id": &hmapi.Content{
				Type:  hmapi.MediaTypeHMAPIString,
				Value: t.id,
			},
			"cmd": &hmapi.Content{
				Type:  hmapi.MediaTypeHMAPIString,
				Value: t.cmd.Path,
			},
			"args": &hmapi.Content{
				Type:  hmapi.MediaTypeJSON,
				Value: t.cmd.Args,
			},
		},
	}

	rw.Header().Set("Content-Type", hmapi.MediaTypeJSON.String())
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(&resource)
}

func (t *process) start(rw http.ResponseWriter, r *http.Request) {
	stdoutPipe, err := t.cmd.StdoutPipe()

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	t.stdoutPipe = stdoutPipe

	stderrPipe, err := t.cmd.StderrPipe()

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	t.stderrPipe = stderrPipe

	stdinPipe, err := t.cmd.StdinPipe()

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	t.stdinPipe = stdinPipe

	err = t.cmd.Start()

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	go func() {
		t.cmd.Wait()
		t.cancel()
	}()

	var stdoutbuf bytes.Buffer
	var stderrbuf bytes.Buffer

	go func(t *process) {
		for {
			select {
			case <-t.ctx.Done():
				return
			default:
				break
			}

			t.mu.Lock()

			stdoutn, _ := io.Copy(&stdoutbuf, t.stdoutPipe)
			stderrn, _ := io.Copy(&stderrbuf, t.stderrPipe)

			if stdoutn > 0 {
				for _, reader := range t.stdoutReaders {
					go func(ch chan []byte, b []byte) {
						ch <- b
					}(reader, stdoutbuf.Bytes()[:stdoutn])
				}
			}

			if stderrn > 0 {
				for _, reader := range t.stderrReaders {
					go func(ch chan []byte, b []byte) {
						ch <- b
					}(reader, stderrbuf.Bytes()[:stderrn])
				}
			}

			t.mu.Unlock()
		}
	}(t)
}

func (t *process) stop(rw http.ResponseWriter, r *http.Request) {
	t.cancel()
}

func (t *process) stdin(rw http.ResponseWriter, r *http.Request) {
	ps := t.cmd.ProcessState

	if ps != nil && ps.Exited() {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("cannot supply stdin process has already exited"))
		return
	}

	form := multipart.NewReader(r.Body, hmapi.MultipartFormDataBoundry)
	data, err := form.NextPart()

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	if data.FormName() != "data" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("field 'data' not supplied"))
		return
	}

	close := rw.(http.CloseNotifier).CloseNotify()

	buf := make([]byte, 250000)
	chdata := make(chan []byte)
	cherr := make(chan error)

	go func() {
		for {
			n, err := r.Body.Read(buf)

			if n > 0 {
				chdata <- buf[:n]
			}

			if err != nil && err != io.EOF {
				cherr <- err
			}

			if err == io.EOF {
				return
			}
		}
	}()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-close:
			return
		case data := <-chdata:
			t.stdinPipe.Write(data)
		case err := <-cherr:
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
	}
}

func (t *process) stdout(rw http.ResponseWriter, r *http.Request) {
	ps := t.cmd.ProcessState

	if ps != nil && ps.Exited() {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("cannot supply stdout process has already exited"))
		return
	}

	flush := rw.(http.Flusher).Flush
	close := rw.(http.CloseNotifier).CloseNotify()

	t.mu.Lock()
	ch := make(chan []byte)
	t.stdoutReaders = append(t.stdoutReaders, ch)
	t.mu.Unlock()

	rw.Header().Set("Trailer", "Error")
	rw.Header().Set("Content-Type", hmapi.MediaTypeOctetStream.String())
	rw.WriteHeader(http.StatusOK)
	flush()

	defer func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		for a, reader := range t.stdoutReaders {
			if reader == ch {
				t.stdoutReaders = append(t.stdoutReaders[:a], t.stdoutReaders[a+1:]...)
				return
			}
		}
	}()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-close:
			return
		case data := <-ch:
			rw.Write(data)
			flush()
		}
	}
}

func (t *process) stderr(rw http.ResponseWriter, r *http.Request) {
	ps := t.cmd.ProcessState

	if ps != nil && ps.Exited() {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("cannot supply stderr process has already exited"))
		return
	}

	flush := rw.(http.Flusher).Flush
	close := rw.(http.CloseNotifier).CloseNotify()

	t.mu.Lock()
	ch := make(chan []byte)
	t.stderrReaders = append(t.stderrReaders, ch)
	t.mu.Unlock()

	rw.Header().Set("Trailer", "Error")
	rw.Header().Set("Content-Type", hmapi.MediaTypeOctetStream.String())
	rw.WriteHeader(http.StatusOK)
	flush()

	defer func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		for a, reader := range t.stderrReaders {
			if reader == ch {
				t.stderrReaders = append(t.stderrReaders[:a], t.stderrReaders[a+1:]...)
				return
			}
		}
	}()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-close:
			return
		case data := <-ch:
			rw.Write(data)
			flush()
		}
	}
}
