package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
)

func externalNetwork(c *config.ExternalDomainConfig) string {
	network := strings.TrimSpace(c.GetNetwork())
	if network == "" {
		return connect.ProtocolHTTP
	}
	return strings.ToUpper(network)
}

func externalContext(ctx context.Context, c *config.ExternalDomainConfig) context.Context {
	if c == nil {
		return ctx
	}
	pairs := make([]string, 0, 4)
	if auth := strings.TrimSpace(c.GetJwtToken()); auth != "" && !hasRequestValue(ctx, cnst.HTTPHeaderAuthorization, cnst.MetadataGlobalKeyAuthorization) {
		if !strings.HasPrefix(strings.ToLower(auth), strings.ToLower(cnst.HTTPHeaderBearerPrefix)+" ") {
			auth = cnst.HTTPHeaderBearerPrefix + " " + auth
		}
		pairs = append(pairs, cnst.MetadataGlobalKeyAuthorization, auth)
	}
	if namespace := strings.TrimSpace(c.GetNamespace()); namespace != "" && !hasRequestValue(ctx, cnst.HTTPHeaderXNamespace, cnst.MetadataGlobalKeyNamespace) {
		pairs = append(pairs, cnst.MetadataGlobalKeyNamespace, namespace)
	}
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToClientContext(ctx, pairs...)
}

func hasClientMetadata(ctx context.Context, key string) bool {
	md, ok := metadata.FromClientContext(ctx)
	return ok && strings.TrimSpace(md.Get(key)) != ""
}

func hasRequestValue(ctx context.Context, headerKey, metadataKey string) bool {
	if tr, ok := transport.FromServerContext(ctx); ok && strings.TrimSpace(tr.RequestHeader().Get(headerKey)) != "" {
		return true
	}
	return hasClientMetadata(ctx, metadataKey)
}
