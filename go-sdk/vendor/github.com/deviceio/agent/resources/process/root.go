package process

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"

	"github.com/deviceio/hmapi"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Root struct {
	Itemsmu *sync.Mutex
	Items   map[string]*process
}

func (t *Root) Get(rw http.ResponseWriter, r *http.Request) {
	parentPath := r.Header.Get("X-Deviceio-Parent-Path")

	resource := &hmapi.Resource{
		Forms: map[string]*hmapi.Form{
			"create-process": &hmapi.Form{
				Action:  parentPath + "/process",
				Method:  hmapi.POST,
				Enctype: hmapi.MediaTypeMultipartFormData,
				Fields: []*hmapi.FormField{
					&hmapi.FormField{
						Name:     "cmd",
						Type:     hmapi.MediaTypeHMAPIString,
						Required: true,
					},
					&hmapi.FormField{
						Name:     "arg",
						Type:     hmapi.MediaTypeHMAPIString,
						Multiple: true,
					},
				},
			},
		},
		Content: map[string]*hmapi.Content{},
		Links:   map[string]*hmapi.Link{},
	}

	t.Itemsmu.Lock()
	defer t.Itemsmu.Unlock()

	resource.Content["process-list"] = &hmapi.Content{
		Type:  hmapi.MediaTypeJSON,
		Value: t.Items,
	}

	for id := range t.Items {
		resource.Links[id] = &hmapi.Link{
			Href: parentPath + "/process/" + id,
			Type: hmapi.MediaTypeHMAPIResource,
		}
	}

	rw.Header().Set("Content-Type", hmapi.MediaTypeJSON.String())
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(&resource)
}

func (t *Root) CreateProcess(rw http.ResponseWriter, r *http.Request) {
	parentPath := r.Header.Get("X-Deviceio-Parent-Path")
	formReader, err := r.MultipartReader()

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	form, err := formReader.ReadForm(250000)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	cmds, ok := form.Value["cmd"]

	if !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("cmd not supplied"))
		return
	}

	if len(cmds) != 1 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("only one cmd field should be supplied"))
		return
	}

	cmdstr := cmds[0]

	args, ok := form.Value["arg"]

	if !ok {
		args = []string{}
	}

	t.Itemsmu.Lock()
	defer t.Itemsmu.Unlock()

	procid, _ := uuid.NewRandom()

	ctx, cancel := context.WithCancel(context.Background())

	proc := &process{
		id:     procid.String(),
		cmd:    exec.CommandContext(ctx, cmdstr, args...),
		mu:     &sync.Mutex{},
		ctx:    ctx,
		cancel: cancel,
	}

	t.Items[procid.String()] = proc

	rw.Header().Set("Location", fmt.Sprintf(
		"%v/%v/%v",
		parentPath,
		"process",
		procid.String(),
	))

	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(""))
}

func (t *Root) GetProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.get(rw, r)
}

func (t *Root) DeleteProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.cancel()
	delete(t.Items, vars["proc-id"])

	rw.WriteHeader(http.StatusOK)
}

func (t *Root) StartProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.start(rw, r)
}

func (t *Root) StopProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.stop(rw, r)
}

func (t *Root) StdinProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.stdin(rw, r)
}

func (t *Root) StdoutProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.stdout(rw, r)
}

func (t *Root) StderrProcess(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	proc := t.findProcessItem(vars["proc-id"])

	if proc == nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(""))
		return
	}

	proc.stderr(rw, r)
}

func (t *Root) findProcessItem(processid string) *process {
	t.Itemsmu.Lock()
	defer t.Itemsmu.Unlock()

	proc, ok := t.Items[processid]

	if !ok {
		return nil
	}

	return proc
}
