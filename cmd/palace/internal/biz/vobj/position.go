package vobj

// Position System unified position.
//
//go:generate stringer -type=Position -linecomment -output=position.string.go
type Position int8

const (
	PositionUnknown    Position = iota // unknown
	PositionSuperAdmin                 // super_admin
	PositionAdmin                      // admin
	PositionUser                       // user
	PositionGuest                      // guest
)

// IsAdminOrSuperAdmin Is it admin or super admin
func (i Position) IsAdminOrSuperAdmin() bool {
	return i == PositionAdmin || i == PositionSuperAdmin
}

// GT Determine if it is greater than or equal to.
func (i Position) GT(j Position) bool {
	return !i.IsUnknown() && i < j
}
