package bo

import (
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateDatasourceParams struct {
		// 数据源名称
		Name string `json:"name"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
		// 数据源地址
		Endpoint string `json:"endpoint"`
		// 状态
		Status vobj.Status `json:"status"`
		// 描述
		Remark string `json:"remark"`
		// 数据源配置(json 字符串)
		Config string `json:"config"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	QueryDatasourceListParams struct {
		// 分页, 不传不分页
		Page types.Pagination `json:"page"`
		// 关键字
		Keyword string `json:"keyword"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
		// 状态
		Status vobj.Status `json:"status"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	UpdateDatasourceBaseInfoParams struct {
		ID uint32 `json:"id"`
		// 数据源名称
		Name string `json:"name"`
		// 状态
		Status vobj.Status `json:"status"`
		// 描述
		Remark string `json:"remark"`
	}

	UpdateDatasourceConfigParams struct {
		ID uint32 `json:"id"`
		// 数据源配置(json 字符串)
		Config string `json:"config"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	DatasourceQueryParams struct {
		DatasourceID uint32 `json:"datasourceID"`
		// 查询语句
		Query string `json:"query"`
		// 步长
		Step uint32 `json:"step"`
		// 时间范围
		TimeRange []string `json:"timeRange"`

		// 数据源
		*bizmodel.Datasource `json:"datasource"`
	}

	DatasourceQueryData struct {
		Labels     map[string]string       `json:"labels"`
		ResultType string                  `json:"resultType"`
		Values     []*DatasourceQueryValue `json:"values"`
		Value      *DatasourceQueryValue   `json:"value"`
	}

	DatasourceQueryValue struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}
)
