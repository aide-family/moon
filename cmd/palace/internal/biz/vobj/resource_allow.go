package vobj

// ResourceAllow user status
//
//go:generate stringer -type=ResourceAllow -linecomment -output=resource_allow.string.go
type ResourceAllow int8

const (
	ResourceAllowUnknown    ResourceAllow = iota // unknown
	ResourceAllowNone                            // none
	ResourceAllowSystem                          // system
	ResourceAllowSystemRBAC                      // system-rbac
	ResourceAllowTeam                            // team
	ResourceAllowTeamRBAC                        // team-rbac
	ResourceAllowUser                            // user
)
