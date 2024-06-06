package build

import (
	"fmt"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/types"
)

type Enum interface {
	fmt.Stringer
	GetValue() int
}

func EnumItem(enumType Enum) *api.EnumItem {
	if types.IsNil(enumType) {
		return nil
	}
	return &api.EnumItem{
		Value: int32(enumType.GetValue()),
		Label: enumType.String(),
	}
}
