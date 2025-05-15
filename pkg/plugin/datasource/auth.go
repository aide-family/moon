package datasource

type BasicAuth interface {
	GetUsername() string
	GetPassword() string
}

type TLS interface {
	GetClientCertificate() string
	GetClientKey() string
	GetServerName() string
}
