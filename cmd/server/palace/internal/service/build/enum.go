package build

import (
	"fmt"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
)

type Enum interface {
	fmt.Stringer
	GetValue() int
}

type EnumBuilder struct {
	Enum
}

func NewEnumBuilder(enumType Enum) *EnumBuilder {
	return &EnumBuilder{
		Enum: enumType,
	}
}

func (b *EnumBuilder) ToApi() *api.EnumItem {
	if types.IsNil(b) || types.IsNil(b.Enum) {
		return nil
	}
	return &api.EnumItem{
		Value: int32(b.GetValue()),
		Label: b.String(),
	}
}
