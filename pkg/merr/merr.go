package merr

import kerrors "github.com/go-kratos/kratos/v3/errors"

const (
	ReasonClient          = "CLIENT"
	ReasonUnauthenticated = "UNAUTHENTICATED"
	ReasonForbidden       = "FORBIDDEN"
	ReasonNotFound        = "NOT_FOUND"
	ReasonConflict        = "CONFLICT"
	ReasonInvalid         = "INVALID"
	ReasonServer          = "SERVER"
)

// Client is a caller error.
func Client(message string) error {
	return kerrors.BadRequest(ReasonClient, message)
}

// ClientWithMetadata 是带附加信息的调用方错误。前端用它展示精确的等待时间等。
func ClientWithMetadata(message string, metadata map[string]string) error {
	return kerrors.BadRequest(ReasonClient, message).WithMetadata(metadata)
}

// Unauthenticated means the caller has no accepted credential.
func Unauthenticated(message string) error {
	return kerrors.Unauthorized(ReasonUnauthenticated, message)
}

// Forbidden means the credential does not allow the operation.
func Forbidden(message string) error {
	return kerrors.Forbidden(ReasonForbidden, message)
}

// NotFound means the record does not exist or is already deleted.
func NotFound(message string) error {
	return kerrors.NotFound(ReasonNotFound, message)
}

// Conflict means the write collides with an existing record.
func Conflict(message string) error {
	return kerrors.Conflict(ReasonConflict, message)
}

// Invalid means the request breaks a domain rule.
func Invalid(message string) error {
	return kerrors.BadRequest(ReasonInvalid, message)
}

// Server is an unexpected failure.
func Server(message string) error {
	return kerrors.InternalServer(ReasonServer, message)
}

// IsNotFound reports whether err was built by NotFound.
func IsNotFound(err error) bool {
	return kerrors.Reason(err) == ReasonNotFound
}

// IsConflict reports whether err was built by Conflict.
func IsConflict(err error) bool {
	return kerrors.Reason(err) == ReasonConflict
}

// IsInvalid reports whether err was built by Invalid.
func IsInvalid(err error) bool {
	return kerrors.Reason(err) == ReasonInvalid
}
