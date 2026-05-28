package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/server/middler"
)

func externalNetwork(c *config.ExternalDomainConfig) string {
	network := strings.TrimSpace(c.GetNetwork())
	if network == "" {
		return connect.ProtocolHTTP
	}
	return strings.ToUpper(network)
}

func externalContext(ctx context.Context, c *config.ExternalDomainConfig) context.Context {
	return middler.ExternalClientContext(ctx, c)
}
