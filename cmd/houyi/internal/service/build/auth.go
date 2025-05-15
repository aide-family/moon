package build

import (
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/api/houyi/common"
)

func ToBasicAuth(basicAuth *common.BasicAuth) *do.BasicAuth {
	if basicAuth == nil {
		return nil
	}
	return &do.BasicAuth{
		Username: basicAuth.GetUsername(),
		Password: basicAuth.GetPassword(),
	}
}

func ToTLS(tls *common.TLS) *do.TLS {
	if tls == nil {
		return nil
	}
	return &do.TLS{
		ClientCertificate: tls.GetClientCertificate(),
		ClientKey:         tls.GetClientKey(),
		ServerName:        tls.GetServerName(),
	}
}
