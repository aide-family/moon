package vobj

// Gender 性别
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Gender -linecomment
type Gender int

const (
	// GenderUnknown 未知
	GenderUnknown Gender = iota // 未知

	// GenderMale 男
	GenderMale // 男

	// GenderFemale 女
	GenderFemale // 女
)
