package installation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/deviceio/dsc"
	"github.com/google/uuid"
)

// Install for linux
func Install(org string, huburl string, hubport int, hubSelfSigned bool) error {
	var err error

	config := fmt.Sprintf("/opt/deviceio/agent/%v/config.json", org)
	binary := fmt.Sprintf("/opt/deviceio/agent/%v/bin/deviceio-agent", org)

	m := dsc.NewModule(map[string]dsc.Resource{
		"config": &dsc.File{
			Path: config,
			Mode: 0700,
			ContentFunc: func(f *dsc.File) ([]byte, error) {
				var b *bytes.Buffer

				uuid, err := uuid.NewRandom()

				if err != nil {
					return nil, err
				}

				enc := json.NewEncoder(b)
				enc.SetIndent("", "    ")

				err = enc.Encode(&Config{
					ID:                       uuid.String(),
					Tags:                     []string{},
					TransportHost:            huburl,
					TransportPort:            hubport,
					TransportAllowSelfSigned: hubSelfSigned,
				})

				if err != nil {
					return nil, err
				}

				return b.Bytes(), nil
			},
		},
		"binary": &dsc.File{
			Path: binary,
			Mode: 0700,
			ContentFunc: func(f *dsc.File) ([]byte, error) {
				exe, err := os.Executable()

				if err != nil {
					return nil, err
				}

				return ioutil.ReadFile(exe)
			},
		},
		"binary-chmod": &dsc.Exec{
			Cmd: "chmod",
			Args: []string{
				"+x",
				binary,
			},
			Relation: &dsc.Relation{
				Require: []string{
					"binary",
				},
			},
		},
		"service": &dsc.Service{
			Name: fmt.Sprintf("deviceio-agent-%v", org),
			Path: fmt.Sprintf("/opt/deviceio/agent/%v/bin/deviceio-agent", org),
			Args: []string{
				"service",
				config,
			},
			Started: true,
			Relation: &dsc.Relation{
				Require: []string{
					"binary",
					"binary-chmod",
					"config",
				},
			},
		},
	})

	if err = m.Run(); err != nil {
		return err
	}

	return nil
}
