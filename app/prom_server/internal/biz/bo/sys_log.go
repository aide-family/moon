package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type ListSyslogReq struct {
	Page     Pagination
	Module   vo.Module
	ModuleId uint32
}

type SysLogBo struct {
	ModuleName vo.Module `json:"moduleName"`
	ModuleId   uint32    `json:"moduleId"`
	Content    string    `json:"content"`
	Title      string    `json:"title"`
	Action     vo.Action `json:"action"`
	CreatedAt  int64     `json:"createdAt"`
	Id         uint32    `json:"id"`
	UserId     uint32    `json:"userId"`
	User       *UserBO   `json:"user"`
}

type ChangeLogBo struct {
	Old any `json:"old"`
	New any `json:"new"`
}

func NewChangeLogBo(old, new any) *ChangeLogBo {
	return &ChangeLogBo{
		Old: old,
		New: new,
	}
}

// String json string
func (l *ChangeLogBo) String() string {
	if l == nil {
		return "{}"
	}
	marshal, err := json.Marshal(l)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// String json string
func (l *SysLogBo) String() string {
	if l == nil {
		return "{}"
	}
	marshal, err := json.Marshal(l)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetUser .
func (l *SysLogBo) GetUser() *UserBO {
	if l == nil {
		return nil
	}
	return l.User
}

func (l *SysLogBo) ToApiV1() *api.SysLogV1Item {
	if l == nil {
		return nil
	}
	return &api.SysLogV1Item{
		Title:      l.Title,
		Content:    l.Content,
		ModuleId:   l.ModuleId,
		ModuleName: api.ModuleType(l.ModuleName),
		CreatedAt:  l.CreatedAt,
		User:       l.GetUser().ToApiSelectV1(),
		Action:     l.Action.Value(),
	}
}

func (l *SysLogBo) ToModel() *do.SysLog {
	if l == nil {
		return nil
	}

	return &do.SysLog{
		BaseModel: do.BaseModel{ID: l.Id},
		Module:    l.ModuleName,
		ModuleId:  l.ModuleId,
		Title:     l.Title,
		Content:   l.Content,
		UserId:    l.UserId,
		Action:    l.Action,
	}
}

// SysLogModelToBo .
func SysLogModelToBo(m *do.SysLog) *SysLogBo {
	if m == nil {
		return nil
	}
	return &SysLogBo{
		ModuleName: m.Module,
		ModuleId:   m.ModuleId,
		Content:    m.Content,
		Title:      m.Title,
		CreatedAt:  m.CreatedAt.Unix(),
		Id:         m.ID,
		UserId:     m.UserId,
		User:       UserModelToBO(m.GetUser()),
		Action:     m.Action,
	}
}
