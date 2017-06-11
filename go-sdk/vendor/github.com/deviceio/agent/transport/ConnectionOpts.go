package transport

type ConnectionOpts struct {
	ID                         string
	Tags                       []string
	TransportAllowSelfSigned   bool
	TransportDisableKeyPinning bool
	TransportHost              string
	TransportPasscodeHash      string
	TransportPasscodeSalt      string
	TransportPort              int
}
