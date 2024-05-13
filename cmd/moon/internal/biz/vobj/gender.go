package vobj

// Gender 性别
//
//go:generate stringer -type=Gender -linecomment
type Gender int

const (
	GenderUnknown Gender = iota // 未知

	GenderMale // 男

	GenderFemale // 女
)
