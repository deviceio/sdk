package sdk

type Device interface {
	Filesystem() DeviceFilesystem
	System() DeviceSystem
	Network() DeviceNetwork
}

type device struct {
	id     string
	client *client
}

func (t *device) Filesystem() DeviceFilesystem {
	return &deviceFilesystem{
		device: t,
	}
}

func (t *device) System() DeviceSystem {
	return nil
}

func (t *device) Network() DeviceNetwork {
	return nil
}
