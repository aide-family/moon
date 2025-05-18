package datasource

type BasicAuth interface {
	GetUsername() string
	GetPassword() string
}

type TLS interface {
	GetClientCert() string
	GetClientKey() string
	GetServerName() string
}
