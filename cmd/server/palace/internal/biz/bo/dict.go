package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateDictParams 创建字典请求参数
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
		CSSClass string `json:"css_class"`
		// icon
		Icon string `json:"icon"`
		// 图片
		ImageURL string `json:"image_url"`
		// 状态
		Status       vobj.Status `json:"status"`
		LanguageCode string      `json:"language_code"`
	}

	// UpdateDictParams 更新字典请求参数
	UpdateDictParams struct {
		ID          uint32 `json:"id"`
		UpdateParam CreateDictParams
	}

	// UpdateDictStatusParams 更新字典状态请求参数
	UpdateDictStatusParams struct {
		IDs    []uint32    `json:"ids"`
		Status vobj.Status `json:"status"`
	}

	// QueryDictListParams 查询字典列表请求参数
	QueryDictListParams struct {
		Keyword  string           `json:"keyword"`
		Page     types.Pagination `json:"page"`
		Status   vobj.Status      `json:"status"`
		DictType vobj.DictType    `json:"dict_type"`
	}
)
