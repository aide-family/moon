package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateMqDatasourceParams 创建消息mq数据源参数
	CreateMqDatasourceParams struct {
		Name string `json:"name"`
		//  数据源类型
		DatasourceType vobj.DatasourceType `json:"datasourceType"`
		StorageType    vobj.StorageType    `json:"storageType"`
		// 数据源地址
		Endpoint string `json:"endpoint"`
		// 状态
		Status vobj.Status `json:"status"`
		// 描述
		Remark string `json:"remark"`
		// 数据源配置(json 字符串)
		Config map[string]string `json:"config"`
	}

	// UpdateMqDatasourceParams 更新消息mq数据源参数
	UpdateMqDatasourceParams struct {
		ID          uint32                    `json:"id"`
		UpdateParam *CreateMqDatasourceParams `json:"updateParam"`
	}

	// QueryMqDatasourceListParams 查询消息mq数据源列表参数
	QueryMqDatasourceListParams struct {
		// 分页, 不传不分页
		Page types.Pagination `json:"page"`
		// 关键字
		Keyword string `json:"keyword"`
		//  数据源类型
		DatasourceType vobj.DatasourceType `json:"datasourceType"`
		StorageType    vobj.StorageType    `json:"storageType"`
		// 状态
		Status vobj.Status `json:"status"`
	}

	UpdateMqDatasourceStatusParams struct {
		ID     uint32      `json:"id"`
		Status vobj.Status `json:"status"`
	}
)
