package sdk

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type DeviceFilesystem interface {
	Read(path string, offset, offsetAt, count int) (DeviceFilesystemReader, error)
}

type DeviceFilesystemReader interface {
	io.Reader
}

type deviceFilesystem struct {
	device *device
}

func (t *deviceFilesystem) Read(path string, offset, offsetAt, count int) (DeviceFilesystemReader, error) {
	deviceResourcePath := fmt.Sprintf("/device/%v/filesystem", t.device.id)

	form, err := t.device.client.hmclient.
		Resource(deviceResourcePath).
		Form("read").
		AddFieldAsString("path", path).
		Submit()

	if err != nil {
		return nil, err
	}

	resp := form.HttpResponse()

	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, &ErrInvalidAPIResponse{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	defer func() {
		if resp.Trailer.Get("Error") != "" {
			os.Stderr.Write([]byte(resp.Trailer.Get("Error")))
			os.Stderr.Sync()
		}
	}()

	buf := make([]byte, 250000)

	if _, err := io.CopyBuffer(os.Stdout, resp.Body, buf); err != nil {
		os.Stderr.Write([]byte(err.Error()))
	}

	return nil, nil
}
