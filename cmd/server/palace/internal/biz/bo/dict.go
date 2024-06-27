package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateDictParams struct {
		// 字典名称
		Name string `json:"name"`
		// 备注
		Remark string `json:"remark"`
		// 字典值
		Value string `json:"value"`
		// 字典类型
		DictType vobj.DictType `json:"dict_type"`
		// 颜色样式
		ColorType string `json:"color_type"`
		// css样式
		CssClass string `json:"css_class"`
		// icon
		Icon string `json:"icon"`
		// 图片
		ImageUrl string `json:"image_url"`
		// 状态
		Status       vobj.Status `json:"status"`
		LanguageCode string      `json:"language_code"`
	}

	UpdateDictParams struct {
		ID          uint32 `json:"id"`
		UpdateParam CreateDictParams
	}

	UpdateDictStatusParams struct {
		IDs    []uint32    `json:"ids"`
		Status vobj.Status `json:"status"`
	}

	QueryDictListParams struct {
		Keyword  string           `json:"keyword"`
		Page     types.Pagination `json:"page"`
		Status   vobj.Status      `json:"status"`
		DictType vobj.DictType    `json:"dict_type"`
	}
)
