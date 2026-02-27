// Package cnst provides constants for HTTP headers and security related constants	.
package cnst

const (
	HTTPHeaderAcceptLang     = "Accept-Language"
	HTTHeaderLang            = "Lang"
	HTTHeaderXRequestID      = "X-Request-ID"
	HTTHeaderXRealIP         = "X-Real-IP"
	HTTHeaderXForwardedFor   = "X-Forwarded-For"
	HTTHeaderXForwardedProto = "X-Forwarded-Proto"
	HTTPHeaderAuthorization  = "Authorization"
	HTTPHeaderXNamespace     = "X-Namespace"
)

const (
	HTTPHeaderBearerPrefix    = "Bearer"
	HTTPHeaderContextTypeJSON = "application/json; charset=utf-8"
	HTTPHeaderContextTypeForm = "application/x-www-form-urlencoded"
	HTTPHeaderContextTypeText = "text/plain; charset=utf-8"
)
