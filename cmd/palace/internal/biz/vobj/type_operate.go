package vobj

// OperateType operate type
//
//go:generate stringer -type=OperateType -linecomment -output=type_operate.string.go
type OperateType int8

const (
	OperateTypeUnknown OperateType = iota // unknown
	OperateTypeQuery                      // query
	OperateTypeAdd                        // add
	OperateTypeUpdate                     // update
	OperateTypeDelete                     // delete
	OperateTypeLogin                      // login
	OperateTypeLogout                     // logout
	OperateTypeExport                     // export
	OperateTypeImport                     // import
)
