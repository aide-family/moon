package merr

import kerrors "github.com/go-kratos/kratos/v2/errors"

const (
	ReasonClient          = "CLIENT"
	ReasonUnauthenticated = "UNAUTHENTICATED"
	ReasonForbidden       = "FORBIDDEN"
	ReasonServer          = "SERVER"
)

// Client is a caller error.
func Client(message string) error {
	return kerrors.BadRequest(ReasonClient, message)
}

// Unauthenticated means the caller has no accepted credential.
func Unauthenticated(message string) error {
	return kerrors.Unauthorized(ReasonUnauthenticated, message)
}

// Forbidden means the credential does not allow the operation.
func Forbidden(message string) error {
	return kerrors.Forbidden(ReasonForbidden, message)
}

// Server is an unexpected failure.
func Server(message string) error {
	return kerrors.InternalServer(ReasonServer, message)
}
