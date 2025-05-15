package vobj

// Role System unified role.
//
//go:generate stringer -type=Role -linecomment -output=role.string.go
type Role int8

const (
	RoleUnknown    Role = iota // unknown
	RoleSuperAdmin             // super_admin
	RoleAdmin                  // admin
	RoleUser                   // user
	RoleGuest                  // guest
)

// IsAdminOrSuperAdmin Is it admin or super admin
func (i Role) IsAdminOrSuperAdmin() bool {
	return i == RoleAdmin || i == RoleSuperAdmin
}

// GT Determine if it is greater than or equal to.
func (i Role) GT(j Role) bool {
	return !i.IsUnknown() && i < j
}
