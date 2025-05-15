package vobj

// Gender gender
//
//go:generate stringer -type=Gender -linecomment -output=gender.string.go
type Gender int8

const (
	GenderUnknown Gender = iota // unknown
	GenderMale                  // male
	GenderFemale                // female
)
