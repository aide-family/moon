package middler

import (
	"context"
	"strconv"
	"strings"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
)

// ExternalClientContext attaches outbound auth and namespace for automated service-to-service calls.
// Namespace from context (e.g. alert labels) takes precedence over static config.
func ExternalClientContext(ctx context.Context, c *config.ExternalDomainConfig) context.Context {
	if c == nil {
		return appendOutboundMetadata(ctx, nil)
	}
	pairs := make([]string, 0, 8)
	if serviceKey := strings.TrimSpace(c.GetServiceKey()); serviceKey != "" &&
		!hasOutboundValue(ctx, cnst.HTTPHeaderAuthorization, cnst.MetadataGlobalKeyAuthorization) {
		ctx = contextx.WithAuthMode(ctx, contextx.AuthModeServiceKey)
		pairs = append(pairs, cnst.MetadataGlobalKeyAuthorization, FormatServiceKeyAuthorization(serviceKey))
	} else if jwtToken := strings.TrimSpace(c.GetJwtToken()); jwtToken != "" &&
		!hasOutboundValue(ctx, cnst.HTTPHeaderAuthorization, cnst.MetadataGlobalKeyAuthorization) {
		ctx = contextx.WithAuthMode(ctx, contextx.AuthModeJWT)
		if _, _, ok := ParseBearerAuthorization(jwtToken); !ok {
			jwtToken = cnst.HTTPHeaderBearerPrefix + " " + jwtToken
		}
		pairs = append(pairs, cnst.MetadataGlobalKeyAuthorization, jwtToken)
	}
	pairs = appendOutboundNamespacePairs(ctx, pairs, c)
	return appendOutboundMetadata(ctx, pairs)
}

func appendOutboundNamespacePairs(ctx context.Context, pairs []string, c *config.ExternalDomainConfig) []string {
	if hasOutboundValue(ctx, cnst.HTTPHeaderXNamespace, cnst.MetadataGlobalKeyNamespace) {
		return pairs
	}
	if namespaceUID, ok := contextx.TryGetNamespace(ctx); ok {
		return append(pairs, cnst.MetadataGlobalKeyNamespace, strconv.FormatInt(namespaceUID.Int64(), 10))
	}
	if c != nil {
		if namespace := strings.TrimSpace(c.GetNamespace()); namespace != "" {
			return append(pairs, cnst.MetadataGlobalKeyNamespace, namespace)
		}
	}
	return pairs
}

func appendOutboundMetadata(ctx context.Context, pairs []string) context.Context {
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToClientContext(ctx, pairs...)
}

func hasOutboundValue(ctx context.Context, headerKey, metadataKey string) bool {
	if tr, ok := transport.FromServerContext(ctx); ok && strings.TrimSpace(tr.RequestHeader().Get(headerKey)) != "" {
		return true
	}
	if tr, ok := transport.FromClientContext(ctx); ok && strings.TrimSpace(tr.RequestHeader().Get(headerKey)) != "" {
		return true
	}
	if md, ok := metadata.FromClientContext(ctx); ok && strings.TrimSpace(md.Get(metadataKey)) != "" {
		return true
	}
	if md, ok := metadata.FromServerContext(ctx); ok && strings.TrimSpace(md.Get(metadataKey)) != "" {
		return true
	}
	return false
}
