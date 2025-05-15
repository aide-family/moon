package vobj

// SampleMode is the sample mode of the palace.
//
//go:generate stringer -type=SampleMode -linecomment -output=sample_mode.string.go
type SampleMode int8

const (
	SampleModeUnknown SampleMode = iota // unknown
	SampleModeFor                       // for
	SampleModeMax                       // max
	SampleModeMin                       // min
)
