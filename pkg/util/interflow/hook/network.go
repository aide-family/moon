package hook

type Network string

const (
	NetworkHTTP Network = "http"
	NetworkGRPC Network = "grpc"
)

func (n Network) String() string {
	return string(n)
}

// IsHTTP checks whether the network is HTTP.
func (n Network) IsHTTP() bool {
	return n == NetworkHTTP
}

// IsGRPC checks whether the network is GRPC.
func (n Network) IsGRPC() bool {
	return n == NetworkGRPC
}
