package installation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/deviceio/dsc"
	"github.com/google/uuid"
)

// Install for windows
func Install(org string, huburl string, hubport int, hubSelfSigned bool) error {
	var err error

	m := dsc.NewModule(map[string]dsc.Resource{
		"config": &dsc.File{
			Path: fmt.Sprintf("c:/PROGRA~1/deviceio/agent/%v/config.json", org),
			Mode: 0700,
			ContentFunc: func(f *dsc.File) ([]byte, error) {
				var b bytes.Buffer

				uuid, err := uuid.NewRandom()

				if err != nil {
					return nil, err
				}

				enc := json.NewEncoder(&b)
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
			Path: fmt.Sprintf("c:/PROGRA~1/deviceio/agent/%v/bin/deviceio-agent.exe", org),
			Mode: 0700,
			ContentFunc: func(f *dsc.File) ([]byte, error) {
				exe, err := os.Executable()

				if err != nil {
					return nil, err
				}

				return ioutil.ReadFile(exe)
			},
			Relation: &dsc.Relation{
				Require: []string{
					"config",
				},
			},
		},
		"service": &dsc.Service{
			Name: fmt.Sprintf("Deviceio Agent (%v)", org),
			Path: fmt.Sprintf(strings.Replace("c:/PROGRA~1/deviceio/agent/%v/bin/deviceio-agent.exe", "/", "\\", -1), org),
			Args: []string{
				"service",
				fmt.Sprintf("c:/PROGRA~1/deviceio/agent/%v/config.json", org),
			},
			Started: true,
			Relation: &dsc.Relation{
				Require: []string{
					"config",
					"binary",
				},
			},
		},
	})

	if err = m.Run(); err != nil {
		return err
	}

	return nil
}
