package bo

import (
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateDatasourceParams 创建数据源请求参数
	CreateDatasourceParams struct {
		// 数据源名称
		Name string `json:"name"`
		// 数据源类型
		DatasourceType vobj.DatasourceType `json:"datasourceType"`
		// 数据源地址
		Endpoint string `json:"endpoint"`
		// 状态
		Status vobj.Status `json:"status"`
		// 描述
		Remark string `json:"remark"`
		// 数据源配置(json 字符串)
		Config map[string]string `json:"config"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	// QueryDatasourceListParams 查询数据源列表请求参数
	QueryDatasourceListParams struct {
		// 分页, 不传不分页
		Page types.Pagination `json:"page"`
		// 关键字
		Keyword string `json:"keyword"`
		// 数据源类型
		DatasourceType vobj.DatasourceType `json:"datasourceType"`
		// 状态
		Status vobj.Status `json:"status"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	// UpdateDatasourceBaseInfoParams 更新数据源基础信息请求参数
	UpdateDatasourceBaseInfoParams struct {
		ID uint32 `json:"id"`
		// 数据源名称
		Name string `json:"name"`
		// 状态
		Status vobj.Status `json:"status"`
		// 数据源配置(json 字符串)
		ConfigValue string `json:"configValue"`
		// 描述
		Remark         string              `json:"remark"`
		StorageType    vobj.StorageType    `json:"storageType"`
		DatasourceType vobj.DatasourceType `json:"datasourceType"`
	}

	// UpdateDatasourceConfigParams 更新数据源配置请求参数
	UpdateDatasourceConfigParams struct {
		ID uint32 `json:"id"`
		// 数据源配置(json 字符串)
		ConfigValue string `json:"configValue"`
		// 数据源类型
		Type vobj.DatasourceType `json:"type"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	// DatasourceQueryParams 查询数据源请求参数
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

	// MetricQueryData 数据源查询结果
	MetricQueryData struct {
		Labels     map[string]string       `json:"labels"`
		ResultType string                  `json:"resultType"`
		Values     []*DatasourceQueryValue `json:"values"`
		Value      *DatasourceQueryValue   `json:"value"`
	}

	// DatasourceQueryValue 数据源查询结果值
	DatasourceQueryValue struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}
)
