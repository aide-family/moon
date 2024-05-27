package vobj

import (
	"database/sql/driver"
)

//var _ driver.Valuer = Gender(0)

// Gender 性别
//
//go:generate stringer -type=Gender -linecomment
type Gender int

func (i Gender) Value() (driver.Value, error) {
	return i, nil
}

const (
	GenderUnknown Gender = iota // 未知

	GenderMale // 男

	GenderFemale // 女
)
