package team

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

var _ do.OperateLog = (*OperateLog)(nil)

const tableNameOperateLog = "team_operate_logs"

type OperateLog struct {
	do.TeamModel
	OperateType     vobj.OperateType    `gorm:"column:type;type:tinyint(2);not null;comment:操作类型" json:"operateType"`
	OperateModule   vobj.ResourceModule `gorm:"column:module;type:tinyint(2);not null;comment:资源模块" json:"operateModule"`
	OperateDataID   uint32              `gorm:"column:data_id;type:int unsigned;not null;comment:操作数据id" json:"operateDataID"`
	OperateDataName string              `gorm:"column:data_name;type:varchar(255);not null;comment:操作数据名称" json:"operateDataName"`
	Title           string              `gorm:"column:title;type:varchar(255);not null;comment:标题" json:"title"`
	Before          string              `gorm:"column:before;type:text;not null;comment:操作前" json:"before"`
	After           string              `gorm:"column:after;type:text;not null;comment:操作后" json:"after"`
	IP              string              `gorm:"column:ip;type:varchar(128);not null;comment:ip" json:"ip"`
}

func (o *OperateLog) GetOperateType() vobj.OperateType {
	if o == nil {
		return vobj.OperateTypeUnknown
	}
	return o.OperateType
}

func (o *OperateLog) GetOperateModule() vobj.ResourceModule {
	if o == nil {
		return vobj.ResourceModuleUnknown
	}
	return o.OperateModule
}

func (o *OperateLog) GetOperateDataID() uint32 {
	if o == nil {
		return 0
	}
	return o.OperateDataID
}

func (o *OperateLog) GetOperateDataName() string {
	if o == nil {
		return ""
	}
	return o.OperateDataName
}

func (o *OperateLog) GetTitle() string {
	if o == nil {
		return ""
	}
	return o.Title
}

func (o *OperateLog) GetBefore() string {
	if o == nil {
		return ""
	}
	return o.Before
}

func (o *OperateLog) GetAfter() string {
	if o == nil {
		return ""
	}
	return o.After
}

func (o *OperateLog) GetIP() string {
	if o == nil {
		return ""
	}
	return o.IP
}

func (o *OperateLog) TableName() string {
	return tableNameOperateLog
}
