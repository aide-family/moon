package model

import (
	"fmt"
)

var _ fmt.Stringer = (*Category)(nil)
var _ fmt.Stringer = (*Status)(nil)

type (
	Category int
	Status   int
)

const (
	Enable Status = iota + 1
	// Disable 关闭
	Disable
)

const (
	// CategoryPromLabel 用于标签
	CategoryPromLabel Category = iota + 1
	// CategoryPromAnnotation 用于注释
	CategoryPromAnnotation
	// CategoryPromStrategy 用于策略
	CategoryPromStrategy
	// CategoryPromStrategyGroup 用于策略组
	CategoryPromStrategyGroup
)

const unknown = "未知"

func (s Status) String() string {
	switch s {
	case Enable:
		return "启用"
	case Disable:
		return "禁用"
	default:
		return unknown
	}
}

func (s Status) Value() int {
	return int(s)
}

func (c Category) String() string {
	switch c {
	case CategoryPromLabel:
		return "标签"
	case CategoryPromAnnotation:
		return "注释"
	case CategoryPromStrategy:
		return "策略"
	case CategoryPromStrategyGroup:
		return "策略组"
	default:
		return unknown
	}
}

func (c Category) Value() int {
	return int(c)
}

func (c Category) IsUnknown() bool {
	return c == 0
}
