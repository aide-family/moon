package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToSelectItem(item bo.SelectItem) *common.SelectItem {
	if validate.IsNil(item) {
		return nil
	}
	return &common.SelectItem{
		Value:    item.GetValue(),
		Label:    item.GetLabel(),
		Disabled: item.GetDisabled(),
		Extra:    ToSelectItemExtra(item.GetExtra()),
	}
}

func ToSelectItemExtra(extra bo.SelectItemExtra) *common.SelectItem_Extra {
	if validate.IsNil(extra) {
		return nil
	}
	return &common.SelectItem_Extra{
		Remark: extra.GetRemark(),
		Icon:   extra.GetIcon(),
		Color:  extra.GetColor(),
	}
}

func ToSelectItems(items []bo.SelectItem) []*common.SelectItem {
	return slices.MapFilter(items, func(item bo.SelectItem) (*common.SelectItem, bool) {
		if validate.IsNil(item) {
			return nil, false
		}
		return ToSelectItem(item), true
	})
}
