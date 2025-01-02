package data

import (
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gorm/clause"
)

func initMainDatabase(d *Data) error {
	if env.Env() != "dev" {
		return nil
	}
	if err := d.mainDB.AutoMigrate(model.Models()...); err != nil {
		return err
	}

	if err := query.Use(d.mainDB).SysDict.Clauses(clause.OnConflict{DoNothing: true}).Create(defaultDictList...); err != nil {
		return err
	}

	// 创建资源列表
	if err := query.Use(d.mainDB).SysAPI.Clauses(clause.OnConflict{DoNothing: true}).Create(resourceList...); err != nil {
		return err
	}

	// 创建发送模板
	if err := query.Use(d.mainDB).SysSendTemplate.Clauses(clause.OnConflict{DoNothing: true}).Create(sendTemplateList...); err != nil {
		return err
	}

	return nil
}

// syncBizDatabase 同步业务模型到各个团队， 保证数据一致性
func syncBizDatabase(d *Data) error {
	if env.Env() != "dev" {
		return nil
	}
	// 获取所有团队
	teams, err := query.Use(d.mainDB).SysTeam.Find()
	if !types.IsNil(err) {
		return err
	}
	mainQuery := query.Use(d.mainDB)
	sysApis, err := mainQuery.SysAPI.Find()
	if !types.IsNil(err) {
		return err
	}

	sysDict, err := mainQuery.SysDict.Find()
	if !types.IsNil(err) {
		return err
	}

	sendTemplates, err := mainQuery.SysSendTemplate.Find()
	if !types.IsNil(err) {
		return err
	}

	teamApis := types.SliceToWithFilter(sysApis, func(apiItem *model.SysAPI) (*bizmodel.SysTeamAPI, bool) {
		return &bizmodel.SysTeamAPI{
			Name:   apiItem.Name,
			Path:   apiItem.Path,
			Status: apiItem.Status,
			Remark: apiItem.Remark,
			Module: apiItem.Module,
			Domain: apiItem.Domain,
		}, true
	})

	dictList := types.SliceToWithFilter(sysDict, func(dictItem *model.SysDict) (*bizmodel.SysDict, bool) {
		return &bizmodel.SysDict{
			Name:         dictItem.Name,
			Value:        dictItem.Value,
			DictType:     dictItem.DictType,
			ColorType:    dictItem.ColorType,
			CSSClass:     dictItem.CSSClass,
			Icon:         dictItem.Icon,
			ImageURL:     dictItem.ImageURL,
			Status:       dictItem.Status,
			LanguageCode: dictItem.LanguageCode,
			Remark:       dictItem.Remark,
		}, true
	})

	sendTemplatesList := types.SliceToWithFilter(sendTemplates, func(item *model.SysSendTemplate) (*bizmodel.SysSendTemplate, bool) {
		return &bizmodel.SysSendTemplate{
			Name:     item.Name,
			Content:  item.Content,
			Status:   item.Status,
			Remark:   item.Remark,
			SendType: item.SendType,
		}, true
	})

	for _, team := range teams {
		// 获取团队业务库连接
		db, err := d.GetBizGormDB(team.ID)
		if err != nil {
			return err
		}
		if err = db.AutoMigrate(bizmodel.Models()...); err != nil {
			return err
		}
		// 同步实时告警数据库
		alarmDB, err := d.GetAlarmGormDB(team.ID)
		if err != nil {
			return err
		}
		if err = alarmDB.AutoMigrate(alarmmodel.Models()...); err != nil {
			return err
		}
		if len(dictList) > 0 {
			if err = bizquery.Use(db).SysDict.Clauses(clause.OnConflict{DoNothing: true}).Create(dictList...); !types.IsNil(err) {
				return err
			}
		}
		if err := bizquery.Use(db).SysTeamAPI.Clauses(clause.OnConflict{DoNothing: true}).Create(teamApis...); !types.IsNil(err) {
			return err
		}
		teamMember := &bizmodel.SysTeamMember{
			UserID: team.GetCreatorID(),
			Status: vobj.StatusEnable,
			Role:   vobj.RoleSuperAdmin,
		}
		// 把创建人同步到团队成员表
		if err := bizquery.Use(db).SysTeamMember.Clauses(clause.OnConflict{DoNothing: true}).Create(teamMember); !types.IsNil(err) {
			return err
		}

		if len(sendTemplatesList) > 0 {
			if err := bizquery.Use(db).SysSendTemplate.Clauses(clause.OnConflict{DoNothing: true}).Create(sendTemplatesList...); err != nil {
				return err
			}
		}
	}
	return nil
}

// 创建默认字典
var defaultDictList = []*model.SysDict{
	{
		AllFieldModel: model.AllFieldModel{ID: 1},
		Name:          "一级告警",
		Value:         "1",
		DictType:      vobj.DictTypeAlarmLevel,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
		CSSClass:      "#a8071a",
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 2},
		Name:          "二级告警",
		Value:         "2",
		DictType:      vobj.DictTypeAlarmLevel,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
		CSSClass:      "#ff9c6e",
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 3},
		Name:          "三级告警",
		Value:         "3",
		DictType:      vobj.DictTypeAlarmLevel,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
		CSSClass:      "#fa8c16",
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 4},
		Name:          "四级告警",
		Value:         "4",
		DictType:      vobj.DictTypeAlarmLevel,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
		CSSClass:      "#d48806",
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 5},
		Name:          "五级告警",
		Value:         "5",
		DictType:      vobj.DictTypeAlarmLevel,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
		CSSClass:      "#d4b106",
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 6},
		Name:          "实时告警",
		Value:         "real-time-alarm-page",
		DictType:      vobj.DictTypeAlarmPage,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 7},
		Name:          "测试告警",
		Value:         "test-alarm-page",
		DictType:      vobj.DictTypeAlarmPage,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 8},
		Name:          "夜班告警",
		Value:         "night-alarm-page",
		DictType:      vobj.DictTypeAlarmPage,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 9},
		Name:          "白班告警",
		Value:         "white-alarm-page",
		DictType:      vobj.DictTypeAlarmPage,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 10},
		Name:          "系统健康",
		Value:         "white-alarm-page",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 11},
		Name:          "系统异常",
		Value:         "system-exception",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 12},
		Name:          "系统告警",
		Value:         "system-alarm",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 13},
		Name:          "系统资源",
		Value:         "system-resource",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 14},
		Name:          "系统配置",
		Value:         "system-config",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 15},
		Name:          "网络状态",
		Value:         "network-status",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 16},
		Name:          "系统负载",
		Value:         "system-load",
		DictType:      vobj.DictTypeStrategyCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 17},
		Name:          "服务器",
		Value:         "server",
		DictType:      vobj.DictTypeStrategyGroupCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 18},
		Name:          "数据库",
		Value:         "database",
		DictType:      vobj.DictTypeStrategyGroupCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 19},
		Name:          "应用",
		Value:         "application",
		DictType:      vobj.DictTypeStrategyGroupCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 20},
		Name:          "网络",
		Value:         "network",
		DictType:      vobj.DictTypeStrategyGroupCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 21},
		Name:          "存储",
		Value:         "storage",
		DictType:      vobj.DictTypeStrategyGroupCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
	{
		AllFieldModel: model.AllFieldModel{ID: 22},
		Name:          "其他",
		Value:         "other",
		DictType:      vobj.DictTypeStrategyGroupCategory,
		Status:        vobj.StatusEnable,
		LanguageCode:  vobj.LanguageZHCN,
	},
}

var resourceList = []*model.SysAPI{
	{
		Name:   "创建告警组",
		Path:   "/api.admin.alarm.Alarm/CreateAlarmGroup",
		Status: vobj.StatusEnable,
		Remark: "用于统一关联告警通知的人员和hook数据集合， 这里是创建这么一个集合",
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "删除告警组",
		Path:   "/api.admin.alarm.Alarm/DeleteAlarmGroup",
		Status: vobj.StatusEnable,
		Remark: "删除告警组",
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "告警组列表",
		Path:   "/api.admin.alarm.Alarm/ListAlarmGroup",
		Remark: "告警组列表， 用于获取告警组列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 告警组详情
	{
		Name:   "告警组详情",
		Path:   "/api.admin.alarm.Alarm/GetAlarmGroup",
		Remark: "告警组详情， 用于获取告警组详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 修改告警组
	{
		Name:   "修改告警组",
		Path:   "/api.admin.alarm.Alarm/UpdateAlarmGroup",
		Remark: "修改告警组， 用于修改告警组",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 修改告警分组状态
	{
		Name:   "修改告警组状态",
		Path:   "/api.admin.alarm.Alarm/UpdateAlarmGroupStatus",
		Remark: "修改告警组状态， 用于修改告警组状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 告警组下拉列表
	{
		Name:   "告警组下拉列表",
		Path:   "/api.admin.alarm.Alarm/ListAlarmGroupSelect",
		Remark: "告警组下拉列表， 用于获取告警组下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 我的告警组
	{
		Name:   "我的告警组",
		Path:   "/api.admin.alarm.Alarm/MyAlarmGroupList",
		Remark: "我的告警组， 用于获取我的告警组",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 消息测试
	{
		Name:   "消息测试",
		Path:   "/api.admin.alarm.Alarm/MessageTest",
		Remark: "消息测试， 用于测试消息发送",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 数据源管理模块
	// 创建数据源
	{
		Name:   "创建数据源",
		Path:   "/api.admin.datasource.Datasource/CreateDatasource",
		Remark: "创建数据源， 用于创建数据源",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 更新数据源
	{
		Name:   "更新数据源",
		Path:   "/api.admin.datasource.Datasource/UpdateDatasource",
		Remark: "更新数据源， 用于更新数据源",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除数据源
	{
		Name:   "删除数据源",
		Path:   "/api.admin.datasource.Datasource/DeleteDatasource",
		Remark: "删除数据源， 用于删除数据源",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取数据源详情
	{
		Name:   "获取数据源详情",
		Path:   "/api.admin.datasource.Datasource/GetDatasource",
		Remark: "获取数据源详情， 用于获取数据源详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取数据源列表
	{
		Name:   "获取数据源列表",
		Path:   "/api.admin.datasource.Datasource/ListDatasource",
		Remark: "获取数据源列表， 用于获取数据源列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 更新数据源状态
	{
		Name:   "更新数据源状态",
		Path:   "/api.admin.datasource.Datasource/UpdateDatasourceStatus",
		Remark: "更新数据源状态， 用于更新数据源状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 下拉列表
	{
		Name:   "数据源下拉列表",
		Path:   "/api.admin.datasource.Datasource/GetDatasourceSelect",
		Remark: "数据源下拉列表， 用于获取数据源下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 同步数据源元数据
	{
		Name:   "同步数据源元数据",
		Path:   "/api.admin.datasource.Datasource/SyncDatasourceMeta",
		Remark: "同步数据源元数据， 用于同步数据源元数据",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取数据
	{
		Name:   "获取数据",
		Path:   "/api.admin.datasource.Datasource/DatasourceQuery",
		Remark: "获取数据， 用于获取数据",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// metric数据源数据查询模块
	// 更新元数据
	{
		Name:   "更新元数据",
		Path:   "/api.admin.datasource.Metric/UpdateMetric",
		Remark: "更新元数据， 用于更新元数据",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取元数据详情
	{
		Name:   "获取元数据详情",
		Path:   "/api.admin.datasource.Metric/GetMetric",
		Remark: "获取元数据详情， 用于获取元数据详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取元数据列表
	{
		Name:   "获取元数据列表",
		Path:   "/api.admin.datasource.Metric/ListMetric",
		Remark: "获取元数据列表， 用于获取元数据列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取元数据列表（下拉选择接口）
	{
		Name:   "获取元数据列表（下拉选择接口）",
		Path:   "/api.admin.datasource.Metric/SelectMetric",
		Remark: "获取元数据列表（下拉选择接口）， 用于获取元数据列表（下拉选择接口）",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除指标
	{
		Name:   "删除指标",
		Path:   "/api.admin.datasource.Metric/DeleteMetric",
		Remark: "删除指标， 用于删除指标",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 字典服务
	// 创建字典
	{
		Name:   "创建字典",
		Path:   "/api.admin.dict.Dict/CreateDict",
		Remark: "创建字典， 用于创建字典",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 更新字典
	{
		Name:   "更新字典",
		Path:   "/api.admin.dict.Dict/UpdateDict",
		Remark: "更新字典， 用于更新字典",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除字典
	{
		Name:   "删除字典",
		Path:   "/api.admin.dict.Dict/DeleteDict",
		Remark: "删除字典， 用于删除字典",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取字典详情
	{
		Name:   "获取字典详情",
		Path:   "/api.admin.dict.Dict/GetDict",
		Remark: "获取字典详情， 用于获取字典详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取字典列表
	{
		Name:   "获取字典列表",
		Path:   "/api.admin.dict.Dict/ListDict",
		Remark: "获取字典列表， 用于获取字典列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 字典类型列表
	{
		Name:   "字典类型列表",
		Path:   "/api.admin.dict.Dict/ListDictType",
		Remark: "字典类型列表， 用于获取字典类型列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取字典下拉列表
	{
		Name:   "获取字典下拉列表",
		Path:   "/api.admin.dict.Dict/DictSelectList",
		Remark: "获取字典下拉列表， 用于获取字典下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 批量修改字典状态
	{
		Name:   "批量修改字典状态",
		Path:   "/api.admin.dict.Dict/BatchUpdateDictStatus",
		Remark: "批量修改字典状态， 用于批量修改字典状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// hook应用
	// 创建hook
	{
		Name:   "创建hook",
		Path:   "/api.admin.hook.Hook/CreateHook",
		Remark: "创建hook， 用于创建hook",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 更新hook
	{
		Name:   "更新hook",
		Path:   "/api.admin.hook.Hook/UpdateHook",
		Remark: "更新hook， 用于更新hook",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除hook
	{
		Name:   "删除hook",
		Path:   "/api.admin.hook.Hook/DeleteHook",
		Remark: "删除hook， 用于删除hook",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取hook详情
	{
		Name:   "获取hook详情",
		Path:   "/api.admin.hook.Hook/GetHook",
		Remark: "获取hook详情， 用于获取hook详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取hook列表
	{
		Name:   "获取hook列表",
		Path:   "/api.admin.hook.Hook/ListHook",
		Remark: "获取hook列表， 用于获取hook列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 批量修改hook状态
	{
		Name:   "批量修改hook状态",
		Path:   "/api.admin.hook.Hook/UpdateHookStatus",
		Remark: "批量修改hook状态， 用于批量修改hook状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取hook下拉列表
	{
		Name:   "获取hook下拉列表",
		Path:   "/api.admin.hook.Hook/ListHookSelectList",
		Remark: "获取hook下拉列表， 用于获取hook下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 邀请模块
	// 邀请User加入团队
	{
		Name:   "邀请User加入团队",
		Path:   "/api.admin.invite.Invite/InviteUser",
		Remark: "邀请User加入团队， 用于邀请User加入团队",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 更新邀请状态
	{
		Name:   "更新邀请状态",
		Path:   "/api.admin.invite.Invite/UpdateInviteStatus",
		Remark: "更新邀请状态， 用于更新邀请状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除邀请
	{
		Name:   "删除邀请",
		Path:   "/api.admin.invite.Invite/DeleteInvite",
		Remark: "删除邀请， 用于删除邀请",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取邀请详情
	{
		Name:   "获取邀请详情",
		Path:   "/api.admin.invite.Invite/GetInvite",
		Remark: "获取邀请详情， 用于获取邀请详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取当前用户邀请列表
	{
		Name:   "获取当前用户邀请列表",
		Path:   "/api.admin.invite.Invite/UserInviteList",
		Remark: "获取当前用户邀请列表， 用于获取当前用户邀请列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 实时告警模块
	// 获取告警详情
	{
		Name:   "获取告警详情",
		Path:   "/api.admin.realtime.Alarm/GetAlarm",
		Remark: "获取告警详情， 用于获取告警详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取告警列表
	{
		Name:   "获取告警列表",
		Path:   "/api.admin.realtime.Alarm/ListAlarm",
		Remark: "获取告警列表， 用于获取告警列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 告警页面模块
	// 维护个人告警页面信息
	{
		Name:   "维护个人告警页面信息",
		Path:   "/api.admin.realtime.AlarmPageSelf/UpdateAlarmPage",
		Remark: "维护个人告警页面信息， 用于维护个人告警页面信息",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取个人告警页面列表
	{
		Name:   "获取个人告警页面列表",
		Path:   "/api.admin.realtime.AlarmPageSelf/ListAlarmPage",
		Remark: "获取个人告警页面列表， 用于获取个人告警页面列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 实时监控之数据大盘
	// 创建大盘
	{
		Name:   "创建大盘",
		Path:   "/api.admin.realtime.Dashboard/CreateDashboard",
		Remark: "创建大盘， 用于创建大盘",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 更新大盘
	{
		Name:   "更新大盘",
		Path:   "/api.admin.realtime.Dashboard/UpdateDashboard",
		Remark: "更新大盘， 用于更新大盘",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除大盘
	{
		Name:   "删除大盘",
		Path:   "/api.admin.realtime.Dashboard/DeleteDashboard",
		Remark: "删除大盘， 用于删除大盘",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取大盘明细
	{
		Name:   "获取大盘明细",
		Path:   "/api.admin.realtime.Dashboard/GetDashboard",
		Remark: "获取大盘明细， 用于获取大盘明细",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取大盘列表
	{
		Name:   "获取大盘列表",
		Path:   "/api.admin.realtime.Dashboard/ListDashboard",
		Remark: "获取大盘列表， 用于获取大盘列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取大盘下拉列表
	{
		Name:   "获取大盘下拉列表",
		Path:   "/api.admin.realtime.Dashboard/ListDashboardSelect",
		Remark: "获取大盘下拉列表， 用于获取大盘下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 批量修改仪表板状态
	{
		Name:   "批量修改仪表板状态",
		Path:   "/api.admin.realtime.Dashboard/BatchUpdateDashboardStatus",
		Remark: "批量修改仪表板状态， 用于批量修改仪表板状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 系统公共API资源管理模块
	// 获取资源详情
	{
		Name:   "获取资源详情",
		Path:   "/api.admin.resource.Resource/GetResource",
		Remark: "获取资源详情， 用于获取资源详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 获取资源列表
	{
		Name:   "获取资源列表",
		Path:   "/api.admin.resource.Resource/ListResource",
		Remark: "获取资源列表， 用于获取资源列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 批量更新资源状态
	{
		Name:   "批量更新资源状态",
		Path:   "/api.admin.resource.Resource/BatchUpdateResourceStatus",
		Remark: "批量更新资源状态， 用于批量更新资源状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 获取资源下拉列表
	{
		Name:   "获取资源下拉列表",
		Path:   "/api.admin.resource.Resource/GetResourceSelectList",
		Remark: "获取资源下拉列表， 用于获取资源下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 策略管理模块
	// 策略组模块
	// 创建策略组
	{
		Name:   "创建策略组",
		Path:   "/api.admin.strategy.Strategy/CreateStrategyGroup",
		Remark: "创建策略组， 用于创建策略组",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除策略组
	{
		Name:   "删除策略组",
		Path:   "/api.admin.strategy.Strategy/DeleteStrategyGroup",
		Remark: "删除策略组， 用于删除策略组",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 策略组列表
	{
		Name:   "策略组列表",
		Path:   "/api.admin.strategy.Strategy/ListStrategyGroup",
		Remark: "策略组列表， 用于策略组列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 策略组详情
	{
		Name:   "策略组详情",
		Path:   "/api.admin.strategy.Strategy/GetStrategyGroup",
		Remark: "策略组详情， 用于策略组详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 修改策略组
	{
		Name:   "修改策略组",
		Path:   "/api.admin.strategy.Strategy/UpdateStrategyGroup",
		Remark: "修改策略组， 用于修改策略组",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 修改策略分组状态
	{
		Name:   "修改策略分组状态",
		Path:   "/api.admin.strategy.Strategy/UpdateStrategyGroupStatus",
		Remark: "修改策略分组状态， 用于修改策略分组状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 策略模块
	// 创建策略
	{
		Name:   "创建策略",
		Path:   "/api.admin.strategy.Strategy/CreateStrategy",
		Remark: "创建策略， 用于创建策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 修改策略
	{
		Name:   "修改策略",
		Path:   "/api.admin.strategy.Strategy/UpdateStrategy",
		Remark: "修改策略， 用于修改策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 修改策略状态
	{
		Name:   "修改策略状态",
		Path:   "/api.admin.strategy.Strategy/UpdateStrategyStatus",
		Remark: "修改策略状态， 用于修改策略状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 删除策略
	{
		Name:   "删除策略",
		Path:   "/api.admin.strategy.Strategy/DeleteStrategy",
		Remark: "删除策略， 用于删除策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取策略
	{
		Name:   "获取策略",
		Path:   "/api.admin.strategy.Strategy/GetStrategy",
		Remark: "获取策略， 用于获取策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 策略列表
	{
		Name:   "策略列表",
		Path:   "/api.admin.strategy.Strategy/ListStrategy",
		Remark: "策略列表， 用于策略列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 复制策略
	{
		Name:   "复制策略",
		Path:   "/api.admin.strategy.Strategy/CopyStrategy",
		Remark: "复制策略， 用于复制策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 策略模版管理
	// 创建策略模版
	{
		Name:   "创建策略模版",
		Path:   "/api.admin.strategy.Template/CreateTemplateStrategy",
		Remark: "创建策略模版， 用于创建策略模版",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 更新策略模版
	{
		Name:   "更新策略模版",
		Path:   "/api.admin.strategy.Template/UpdateTemplateStrategy",
		Remark: "更新策略模版， 用于更新策略模版",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 删除策略模版
	{
		Name:   "删除策略模版",
		Path:   "/api.admin.strategy.Template/DeleteTemplateStrategy",
		Remark: "删除策略模版， 用于删除策略模版",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取策略模版详情
	{
		Name:   "获取策略模版详情",
		Path:   "/api.admin.strategy.Template/GetTemplateStrategy",
		Remark: "获取策略模版详情， 用于获取策略模版详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取策略模版列表
	{
		Name:   "获取策略模版列表",
		Path:   "/api.admin.strategy.Template/ListTemplateStrategy",
		Remark: "获取策略模版列表， 用于获取策略模版列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 更改模板启用状态
	{
		Name:   "更改模板启用状态",
		Path:   "/api.admin.strategy.Template/UpdateTemplateStrategyStatus",
		Remark: "更改模板启用状态， 用于更改模板启用状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 模板校验（返回校验成功的数据或者错误明细）
	{
		Name:   "模板校验",
		Path:   "/api.admin.strategy.Template/ValidateAnnotationsTemplate",
		Remark: "模板校验， 用于模板校验",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 策略订阅模块
	// 当前用户订阅某个策略
	{
		Name:   "当前用户订阅某个策略",
		Path:   "/api.admin.subscriber.Subscriber/UserSubscriberStrategy",
		Remark: "当前用户订阅某个策略， 用于当前用户订阅某个策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 当前取消订阅策略
	{
		Name:   "取消订阅策略",
		Path:   "/api.admin.subscriber.Subscriber/UnSubscriber",
		Remark: "取消订阅策略， 用于取消订阅策略",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 当前用户订阅策略列表
	{
		Name:   "当前用户订阅策略列表",
		Path:   "/api.admin.subscriber.Subscriber/UserSubscriberList",
		Remark: "当前用户订阅策略列表， 用于当前用户订阅策略列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 策略订阅者列表
	{
		Name:   "策略订阅者列表",
		Path:   "/api.admin.subscriber.Subscriber/GetStrategySubscriber",
		Remark: "策略订阅者列表， 用于策略订阅者列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 团队角色管理
	// 创建角色
	{
		Name:   "创建团队角色",
		Path:   "/api.admin.team.Role/CreateRole",
		Remark: "创建团队角色， 用于创建团队角色",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 更新角色
	{
		Name:   "更新团队角色",
		Path:   "/api.admin.team.Role/UpdateRole",
		Remark: "更新团队角色， 用于更新团队角色",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 删除角色
	{
		Name:   "删除团队角色",
		Path:   "/api.admin.team.Role/DeleteRole",
		Remark: "删除团队角色， 用于删除团队角色",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 获取角色详情
	{
		Name:   "获取团队角色详情",
		Path:   "/api.admin.team.Role/GetRole",
		Remark: "获取团队角色详情， 用于获取团队角色详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 获取角色列表
	{
		Name:   "获取团队角色列表",
		Path:   "/api.admin.team.Role/ListRole",
		Remark: "获取团队角色列表， 用于获取团队角色列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 更新角色状态
	{
		Name:   "更新团队角色状态",
		Path:   "/api.admin.team.Role/UpdateRoleStatus",
		Remark: "更新团队角色状态， 用于更新团队角色状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 角色下拉列表
	{
		Name:   "获取团队角色下拉列表",
		Path:   "/api.admin.team.Role/GetRoleSelectList",
		Remark: "获取团队角色下拉列表， 用于获取团队角色下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 多租户下的团队管理
	// 创建团队
	{
		Name:   "创建团队",
		Path:   "/api.admin.team.Team/CreateTeam",
		Remark: "创建团队， 用于创建团队",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 更新团队
	{
		Name:   "更新团队",
		Path:   "/api.admin.team.Team/UpdateTeam",
		Remark: "更新团队， 用于更新团队",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 获取团队详情
	{
		Name:   "获取团队详情",
		Path:   "/api.admin.team.Team/GetTeam",
		Remark: "获取团队详情， 用于获取团队详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取团队列表
	{
		Name:   "获取团队列表",
		Path:   "/api.admin.team.Team/ListTeam",
		Remark: "获取团队列表， 用于获取团队列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 修改团队状态
	{
		Name:   "修改团队状态",
		Path:   "/api.admin.team.Team/UpdateTeamStatus",
		Remark: "修改团队状态， 用于修改团队状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 我的团队， 查看当前用户的团队列表
	{
		Name:   "我的团队",
		Path:   "/api.admin.team.Team/MyTeam",
		Remark: "我的团队， 用于查看当前用户的团队列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 移除团队成员
	{
		Name:   "移除团队成员",
		Path:   "/api.admin.team.Team/RemoveTeamMember",
		Remark: "移除团队成员， 用于移除团队成员",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 设置成管理员
	{
		Name:   "设置成管理员",
		Path:   "/api.admin.team.Team/SetTeamAdmin",
		Remark: "设置成管理员， 用于设置成管理员",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 移除团队管理员
	{
		Name:   "移除团队管理员",
		Path:   "/api.admin.team.Team/RemoveTeamAdmin",
		Remark: "移除团队管理员， 用于移除团队管理员",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 设置成员角色
	{
		Name:   "设置成员角色",
		Path:   "/api.admin.team.Team/SetMemberRole",
		Remark: "设置成员角色， 用于设置成员角色",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 获取团队成员列表
	{
		Name:   "获取团队成员列表",
		Path:   "/api.admin.team.Team/ListTeamMember",
		Remark: "获取团队成员列表， 用于获取团队成员列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 移交超级管理员
	{
		Name:   "移交超级管理员",
		Path:   "/api.admin.team.Team/TransferTeamLeader",
		Remark: "移交超级管理员， 用于移交超级管理员",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 设置团队邮件配置
	{
		Name:   "设置团队邮件配置",
		Path:   "/api.admin.team.Team/SetTeamMailConfig",
		Remark: "设置团队邮件配置， 用于设置团队邮件配置",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 更新团队成员状态
	{
		Name:   "更新团队成员状态",
		Path:   "/api.admin.team.Team/UpdateTeamMemberStatus",
		Remark: "更新团队成员状态， 用于更新团队成员状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 团队成员详情
	{
		Name:   "团队成员详情",
		Path:   "/api.admin.team.Team/GetTeamMemberDetail",
		Remark: "团队成员详情， 用于获取团队成员详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowTeam,
	},
	// 用户个人消息模块
	// 删除消息， 用于清除所有通知
	{
		Name:   "删除消息",
		Path:   "/api.admin.user.Message/DeleteMessages",
		Remark: "删除消息， 用于清除所有通知",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取消息列表， 用于获取我的未读消息
	{
		Name:   "获取消息列表",
		Path:   "/api.admin.user.Message/ListMessage",
		Remark: "获取消息列表， 用于获取我的未读消息",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 用户模块
	// 创建用户
	{
		Name:   "创建用户",
		Path:   "/api.admin.user.User/CreateUser",
		Remark: "创建用户， 用于创建用户",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 更新用户
	{
		Name:   "更新用户",
		Path:   "/api.admin.user.User/UpdateUser",
		Remark: "更新用户， 用于更新用户",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 删除用户
	{
		Name:   "删除用户",
		Path:   "/api.admin.user.User/DeleteUser",
		Remark: "删除用户， 用于删除用户",
		Status: vobj.StatusEnable,
	},
	// 获取用户
	{
		Name:   "获取用户",
		Path:   "/api.admin.user.User/GetUser",
		Remark: "获取用户， 用于获取用户",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 获取个人信息
	{
		Name:   "获取个人信息",
		Path:   "/api.admin.user.User/GetUserSelfBasic",
		Remark: "获取用户个人信息",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 列表用户
	{
		Name:   "用户列表",
		Path:   "/api.admin.user.User/ListUser",
		Remark: "用户列表， 用于获取用户列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 批量修改用户状态
	{
		Name:   "批量修改用户状态",
		Path:   "/api.admin.user.User/BatchUpdateUserStatus",
		Remark: "批量修改用户状态， 用于批量修改用户状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 设置用户角色
	{
		Name:   "设置用户角色",
		Path:   "/api.admin.user.User/SetUserRole",
		Remark: "设置用户角色， 用于设置用户角色",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 重置用户密码
	{
		Name:   "重置用户密码",
		Path:   "/api.admin.user.User/ResetUserPassword",
		Remark: "重置用户密码， 用于重置用户密码",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 用户修改密码
	{
		Name:   "用户修改密码",
		Path:   "/api.admin.user.User/ResetUserPasswordBySelf",
		Remark: "用户修改密码， 用于用户修改密码",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取用户下拉列表
	{
		Name:   "获取用户下拉列表",
		Path:   "/api.admin.user.User/GetUserSelectList",
		Remark: "获取用户下拉列表， 用于获取用户下拉列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowSystem,
	},
	// 修改电话号码
	{
		Name:   "修改电话号码",
		Path:   "/api.admin.user.User/UpdateUserPhone",
		Remark: "修改电话号码， 用于修改电话号码",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 修改邮箱
	{
		Name:   "修改邮箱",
		Path:   "/api.admin.user.User/UpdateUserEmail",
		Remark: "修改邮箱， 用于修改邮箱",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 修改头像
	{
		Name:   "修改头像",
		Path:   "/api.admin.user.User/UpdateUserAvatar",
		Remark: "修改头像， 用于修改头像",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 修改基本信息
	{
		Name:   "修改基本信息",
		Path:   "/api.admin.user.User/UpdateUserBaseInfo",
		Remark: "修改基本信息， 用于修改基本信息",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 认证模块
	// 登录
	{
		Name:   "登录",
		Path:   "/api.admin.authorization.Authorization/Login",
		Remark: "登录， 用于登录",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 登出
	{
		Name:   "登出",
		Path:   "/api.admin.authorization.Authorization/Logout",
		Remark: "登出， 用于登出",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 刷新token
	{
		Name:   "刷新token",
		Path:   "/api.admin.authorization.Authorization/RefreshToken",
		Remark: "刷新token， 用于刷新token",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 获取验证码
	{
		Name:   "获取验证码",
		Path:   "/api.admin.authorization.Authorization/Captcha",
		Remark: "获取验证码， 用于获取验证码",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 设置账号邮箱
	{
		Name:   "设置账号邮箱",
		Path:   "/api.admin.authorization.Authorization/SetEmailWithLogin",
		Remark: "设置账号邮箱， 用于设置账号邮箱",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 验证邮箱
	{
		Name:   "验证邮箱",
		Path:   "/api.admin.authorization.Authorization/VerifyEmail",
		Remark: "验证邮箱， 用于验证邮箱",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 确认消息
	{
		Name:   "确认消息",
		Path:   "/api.admin.user.Message/ConfirmMessage",
		Remark: "确认消息， 用于确认消息",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 取消消息
	{
		Name:   "取消消息",
		Path:   "/api.admin.user.Message/CancelMessage",
		Remark: "取消消息， 用于取消消息",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowUser,
	},
	// 告警历史模块
	// 获取告警历史记录
	{
		Name:   "获取告警历史记录",
		Path:   "/api.admin.history.History/GetHistory",
		Remark: "获取告警历史记录， 用于获取告警历史记录",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 获取告警历史记录列表
	{
		Name:   "获取告警历史记录列表",
		Path:   "/api.admin.history.History/ListHistory",
		Remark: "获取告警历史记录列表， 用于获取告警历史记录列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 实时监控获取服务列表
	{
		Name:   "获取rabbit houyi 服务列表",
		Path:   "/api.Server/GetServerList",
		Remark: "获取rabbit houyi 服务列表 用于前台展示",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowNone,
	},
	// 时间引擎规则模块
	{
		Name:   "获取时间引擎规则列表",
		Path:   "/api.admin.alarm.TimeEngineRule/ListTimeEngineRule",
		Remark: "获取时间引擎规则列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "获取时间引擎规则",
		Path:   "/api.admin.alarm.TimeEngineRule/GetTimeEngineRule",
		Remark: "获取时间引擎规则",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "创建时间引擎规则",
		Path:   "/api.admin.alarm.TimeEngineRule/CreateTimeEngineRule",
		Remark: "创建时间引擎规则",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "更新时间引擎规则",
		Path:   "/api.admin.alarm.TimeEngineRule/UpdateTimeEngineRule",
		Remark: "更新时间引擎规则",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "删除时间引擎规则",
		Path:   "/api.admin.alarm.TimeEngineRule/DeleteTimeEngineRule",
		Remark: "删除时间引擎规则",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "批量更新时间引擎规则状态",
		Path:   "/api.admin.alarm.TimeEngineRule/BatchUpdateTimeEngineRuleStatus",
		Remark: "批量更新时间引擎规则状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "创建时间引擎",
		Path:   "/api.admin.alarm.TimeEngineRule/CreateTimeEngine",
		Remark: "创建时间引擎",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "更新时间引擎",
		Path:   "/api.admin.alarm.TimeEngineRule/UpdateTimeEngine",
		Remark: "更新时间引擎",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "删除时间引擎",
		Path:   "/api.admin.alarm.TimeEngineRule/DeleteTimeEngine",
		Remark: "删除时间引擎",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "获取时间引擎",
		Path:   "/api.admin.alarm.TimeEngineRule/GetTimeEngine",
		Remark: "获取时间引擎",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "获取时间引擎列表",
		Path:   "/api.admin.alarm.TimeEngineRule/ListTimeEngine",
		Remark: "获取时间引擎列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "批量更新时间引擎状态",
		Path:   "/api.admin.alarm.TimeEngineRule/BatchUpdateTimeEngineStatus",
		Remark: "批量更新时间引擎状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	// 发送模板相关
	{
		Name:   "创建发送模板",
		Path:   "/api.admin.template.SendTemplate/CreateSendTemplate",
		Remark: "创建发送模板",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "删除发送模板",
		Path:   "/api.admin.template.SendTemplate/DeleteSendTemplate",
		Remark: "删除发送模板",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "获取发送模板详情",
		Path:   "/api.admin.template.SendTemplate/GetSendTemplate",
		Remark: "获取发送模板详情",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "发送模板列表",
		Path:   "/api.admin.template.SendTemplate/ListSendTemplate",
		Remark: "发送模板列表",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "更新发送模板",
		Path:   "/api.admin.template.SendTemplate/UpdateSendTemplate",
		Remark: "更新发送模板",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
	{
		Name:   "更新发送模板状态",
		Path:   "/api.admin.template.SendTemplate/UpdateStatus",
		Remark: "更新发送模板状态",
		Status: vobj.StatusEnable,
		Allow:  vobj.AllowRBAC,
	},
}

// 发送模板相关
var sendTemplateList = []*model.SysSendTemplate{
	{
		Name:     "邮箱-监控告警模板",
		Content:  "    <h1>监控告警</h1>\n    <p>{{ .annotations.summary }}</p>\n    <p>{{ .annotations.description }}</p>\n    <p>时间: {{ .startsAt }} 至 {{ .endsAt }}</p>",
		SendType: vobj.AlarmSendTypeEmail,
		Status:   vobj.StatusEnable,
		Remark:   "系统邮箱模板",
	},
	{
		Name:     "钉钉-监控告警模板",
		Content:  "    {{- $status := .status -}}\n    {{- $labels := .labels -}}\n    {{- $annotations := .annotations -}}\n    \n    {\n        \"msgtype\": \"markdown\",\n        \"markdown\": {\n          \"title\": \"平台状态通知\",\n          \"text\": \"### {{if eq $status `resolved`}}✅ 告警已恢复{{else}}🚨 紧急告警通知{{end}}\\n\\n  \\n**时间**: `{{ .startsAt }}` 至 `{{ .endsAt }}`  \\n\\n<hr/>\\n\\n**摘要**:  \\n`{{ $annotations.summary }}`  \\n\\n**描述**:  \\n`{{ $annotations.description }}`  \\n\\n<hr/>\\n\\n**标签**:  \\n- **数据源 ID**: {{ index $labels \"__moon__datasource_id__\" }}  \\n- **数据源 URL**: [链接]({{ index $labels \"__moon__datasource_url__\" }})  \\n- **级别 ID**: {{ index $labels \"__moon__level_id__\" }}  \\n- **策略 ID**: {{ index $labels \"__moon__strategy_id__\" }}  \\n- **团队 ID**: {{ index $labels \"__moon__team_id__\" }}  \\n- **实例**: `{{ index $labels \"instance\" }}`  \\n- **IP**: `{{ index $labels \"ip\" }}`  \\n- **作业**: `{{ index $labels \"job\" }}`  \\n\\n<hr/>\\n\\n请根据以上信息进行后续处理！\"\n        }\n    }",
		SendType: vobj.AlarmSendTypeDingTalk,
		Status:   vobj.StatusEnable,
		Remark:   "系统钉钉模板",
	},
	{
		Name:     "飞书-监控告警模板",
		Content:  "    {\n        \"msg_type\": \"interactive\",\n        \"card\": {\n            \"config\": {\n                \"wide_screen_mode\": true\n            },\n            \"header\": {\n                \"title\": {\n                    \"tag\": \"plain_text\",\n                    \"content\": \"{{if eq .status `resolved`}}✅ 告警已恢复{{else}}🚨 紧急告警通知{{end}}\"\n                },\n                \"template\": \"{{if eq .status `resolved`}}green{{else}}red{{end}}\"\n            },\n            \"elements\": [\n                {\n                    \"tag\": \"div\",\n                    \"fields\": [\n                        {\n                            \"is_short\": false,\n                            \"text\": {\n                                \"tag\": \"lark_md\",\n                                \"content\": \"**🔍 告警摘要**\\n{{.annotations.summary}}\"\n                            }\n                        },\n                        {\n                            \"is_short\": false,\n                            \"text\": {\n                                \"tag\": \"lark_md\",\n                                \"content\": \"**🔍 告警描述**\\n{{.annotations.description}}\"\n                            }\n                        },\n                        {\n                            \"is_short\": false,\n                            \"text\": {\n                                \"tag\": \"lark_md\",\n                                \"content\": \"**🕒 开始时间**\\n{{.startsAt}}\"\n                            }\n                        },\n                        {\n                            \"is_short\": false,\n                            \"text\": {\n                                \"tag\": \"lark_md\",\n                                \"content\": \"{{if eq .status `resolved`}}**🕒 恢复时间**\\n{{.endsAt}}{{end}}\"\n                            }\n                        }\n                    ]\n                },\n                {\n                    \"tag\": \"hr\"\n                },\n                {\n                  \"tag\": \"div\",\n                  \"fields\": [\n                    {\n                      \"is_short\": true,\n                      \"text\": {\n                        \"tag\": \"lark_md\",\n                        \"content\": \"**❗ 状态**\\n<font color=\\\"warning\\\">{{.status}}</font>\"\n                      }\n                    },\n                    {\n                      \"is_short\": true,\n                      \"text\": {\n                        \"tag\": \"lark_md\",\n                        \"content\": \"**🌐 数据源**\\n[点击查看]({{.labels.__moon__datasource_url__}})\"\n                      }\n                    },\n                    {\n                      \"is_short\": true,\n                      \"text\": {\n                        \"tag\": \"lark_md\",\n                        \"content\": \"**📛 告警名称**\\n<font color=\\\"info\\\">{{.labels.__name__}}</font>\"\n                      }\n                    },\n                    {\n                      \"is_short\": true,\n                      \"text\": {\n                        \"tag\": \"lark_md\",\n                        \"content\": \"**🔗 策略 ID**\\n{{.labels.__moon__strategy_id__}}\"\n                      }\n                    },\n                    {\n                      \"is_short\": true,\n                      \"text\": {\n                        \"tag\": \"lark_md\",\n                        \"content\": \"**🆔 团队 ID**\\n{{.labels.__moon__team_id__}}\"\n                      }\n                    },\n                    {\n                      \"is_short\": true,\n                      \"text\": {\n                        \"tag\": \"lark_md\",\n                        \"content\": \"**💻 IP 地址**\\n{{.labels.ip}}\"\n                      }\n                    }\n                  ]\n                },\n                {\n                    \"tag\": \"hr\"\n                },\n                {\n                    \"tag\": \"action\",\n                    \"actions\": [\n                        {\n                            \"tag\": \"button\",\n                            \"text\": {\n                                \"tag\": \"lark_md\",\n                                \"content\": \"📄 查看详情\"\n                            },\n                            \"url\": \"{{.labels.__moon__datasource_url__}}\",\n                            \"type\": \"primary\"\n                        }\n                    ]\n                }\n            ]\n        }\n    }",
		SendType: vobj.AlarmSendTypeFeiShu,
		Status:   vobj.StatusEnable,
		Remark:   "系统飞书模板",
	},
	{
		Name:     "企业微信-监控告警模板",
		Content:  "    {\n        \"msgtype\": \"markdown\",\n        \"markdown\": {\n          \"content\": \"### {{if eq .status `resolved`}}✅ 告警已恢复{{else}}🚨 紧急告警通知{{end}}\\n\\n {{ .annotations }}\"\n        }\n    }",
		SendType: vobj.AlarmSendTypeWechat,
		Status:   vobj.StatusEnable,
		Remark:   "企业微信告警模板",
	},
}
