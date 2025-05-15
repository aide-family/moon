package do

type TLS struct {
	ServerName        string `json:"serverName,omitempty"`
	ClientCertificate string `json:"clientCertificate,omitempty"`
	ClientKey         string `json:"clientKey,omitempty"`
}

func (t *TLS) GetClientCertificate() string {
	if t == nil {
		return ""
	}
	return t.ClientCertificate
}

func (t *TLS) GetClientKey() string {
	if t == nil {
		return ""
	}
	return t.ClientKey
}

func (t *TLS) GetServerName() string {
	if t == nil {
		return ""
	}
	return t.ServerName
}

type BasicAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (b *BasicAuth) GetUsername() string {
	if b == nil {
		return ""
	}
	return b.Username
}

func (b *BasicAuth) GetPassword() string {
	if b == nil {
		return ""
	}
	return b.Password
}
