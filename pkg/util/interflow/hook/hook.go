package hook

type (
	HttpServerConfig interface {
		GetUrl() string
	}

	GrpcServerConfig interface {
		GetEndpoints() []string
	}

	HttpConfig interface {
		GetAgent() HttpServerConfig
		GetServer() HttpServerConfig
	}

	GrpcConfig interface {
		GetAgent() GrpcServerConfig
		GetServer() GrpcServerConfig
	}

	Config interface {
		GetHttp() HttpConfig
		GetGrpc() GrpcConfig
	}
)
