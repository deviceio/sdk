package installation

type Config struct {
	ID                       string
	Tags                     []string
	TransportHost            string
	TransportPort            int
	TransportAllowSelfSigned bool
}
