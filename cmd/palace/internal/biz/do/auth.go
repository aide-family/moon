package do

import "github.com/aide-family/moon/pkg/plugin/datasource"

var _ datasource.BasicAuth = (*BasicAuth)(nil)

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetPassword implements datasource.BasicAuth.
func (b *BasicAuth) GetPassword() string {
	return b.Password
}

// GetUsername implements datasource.BasicAuth.
func (b *BasicAuth) GetUsername() string {
	return b.Username
}

var _ datasource.TLS = (*TLS)(nil)

type TLS struct {
	ServerName string `json:"serverName"`
	ClientCert string `json:"clientCert"`
	ClientKey  string `json:"clientKey"`
	SkipVerify bool   `json:"skipVerify"`
}

// GetClientCertificate implements datasource.TLS.
func (t *TLS) GetClientCert() string {
	return t.ClientCert
}

// GetClientKey implements datasource.TLS.
func (t *TLS) GetClientKey() string {
	return t.ClientKey
}

// GetServerName implements datasource.TLS.
func (t *TLS) GetServerName() string {
	return t.ServerName
}

// GetSkipVerify implements datasource.TLS.
func (t *TLS) GetSkipVerify() bool {
	return t.SkipVerify
}
