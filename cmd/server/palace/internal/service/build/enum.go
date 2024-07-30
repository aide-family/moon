package build

import (
	"fmt"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
)

// Enum 枚举类型统一实现接口
type Enum interface {
	fmt.Stringer
	GetValue() int
}

// EnumBuilder 枚举类型构造器
type EnumBuilder struct {
	Enum
}

// NewEnumBuilder 创建枚举类型构造器
func NewEnumBuilder(enumType Enum) *EnumBuilder {
	return &EnumBuilder{
		Enum: enumType,
	}
}

// ToAPI 转换为api枚举类型
func (b *EnumBuilder) ToAPI() *api.EnumItem {
	if types.IsNil(b) || types.IsNil(b.Enum) {
		return nil
	}
	return &api.EnumItem{
		Value: int32(b.GetValue()),
		Label: b.String(),
	}
}
