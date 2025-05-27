package vobj

// OAuthFrom oauth from
//
//go:generate stringer -type=OAuthFrom -linecomment -output=oauth_from.string.go
type OAuthFrom int8

const (
	OAuthFromUnknown OAuthFrom = iota // unknown
	OAuthFromAdmin                    // admin
	OAuthFromPortal                   // portal
)
