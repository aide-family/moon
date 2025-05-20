package build

import (
	"github.com/aide-family/moon/cmd/houyi/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToBasicAuth(basicAuth *common.BasicAuth) *do.BasicAuth {
	if validate.IsNil(basicAuth) {
		return nil
	}
	return &do.BasicAuth{
		Username: basicAuth.GetUsername(),
		Password: basicAuth.GetPassword(),
	}
}

func ToTLS(tls *common.TLS) *do.TLS {
	if validate.IsNil(tls) {
		return nil
	}
	return &do.TLS{
		ClientCert: tls.GetClientCert(),
		ClientKey:  tls.GetClientKey(),
		ServerName: tls.GetServerName(),
	}
}
