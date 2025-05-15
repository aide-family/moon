package vobj

// HTTPMethod represents an HTTP method.
//
//go:generate stringer -type=HTTPMethod -linecomment -output=http_method.string.go
type HTTPMethod int8

const (
	HTTPMethodUnknown HTTPMethod = iota // Unknown
	HTTPMethodGet                       // GET
	HTTPMethodPost                      // POST
	HTTPMethodPut                       // PUT
	HTTPMethodDelete                    // DELETE
	HTTPMethodHead                      // HEAD
	HTTPMethodOptions                   // OPTIONS
	HTTPMethodPatch                     // PATCH
)
