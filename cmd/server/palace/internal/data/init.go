package data

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gorm/clause"
)

func initMainDatabase(d *Data) error {
	if err := d.mainDB.AutoMigrate(model.Models()...); err != nil {
		return err
	}

	// 创建默认字典
	dictList := []*model.SysDict{
		{
			AllFieldModel: model.AllFieldModel{ID: 1},
			Name:          "一级告警",
			Value:         "1",
			DictType:      vobj.DictTypeAlarmLevel,
			Status:        vobj.StatusEnable,
			LanguageCode:  vobj.LanguageZHCN,
		},
		{
			AllFieldModel: model.AllFieldModel{ID: 2},
			Name:          "二级告警",
			Value:         "2",
			DictType:      vobj.DictTypeAlarmLevel,
			Status:        vobj.StatusEnable,
			LanguageCode:  vobj.LanguageZHCN,
		},
		{
			AllFieldModel: model.AllFieldModel{ID: 3},
			Name:          "三级告警",
			Value:         "3",
			DictType:      vobj.DictTypeAlarmLevel,
			Status:        vobj.StatusEnable,
			LanguageCode:  vobj.LanguageZHCN,
		},
		{
			AllFieldModel: model.AllFieldModel{ID: 4},
			Name:          "四级告警",
			Value:         "4",
			DictType:      vobj.DictTypeAlarmLevel,
			Status:        vobj.StatusEnable,
			LanguageCode:  vobj.LanguageZHCN,
		},
		{
			AllFieldModel: model.AllFieldModel{ID: 5},
			Name:          "五级告警",
			Value:         "5",
			DictType:      vobj.DictTypeAlarmLevel,
			Status:        vobj.StatusEnable,
			LanguageCode:  vobj.LanguageZHCN,
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
	if err := query.Use(d.mainDB).SysDict.Clauses(clause.OnConflict{DoNothing: true}).Create(dictList...); err != nil {
		return err
	}

	pass := types.NewPassword(types.MD5("123456" + "3c4d9a0a5a703938dd1d2d46e1c924f9"))
	// 如果没有默认用户，则创建一个默认用户
	user := &model.SysUser{
		AllFieldModel: model.AllFieldModel{ID: 1},
		Username:      "admin",
		Nickname:      "超级管理员",
		Password:      pass.String(),
		Email:         "moonio@moon.com",
		Phone:         "18812341234",
		Remark:        "这是个人很懒， 没有设置备注信息",
		Avatar:        "https://img0.baidu.com/it/u=1128422789,3129806361&fm=253&app=120&size=w931&n=0&f=JPEG&fmt=auto?sec=1719766800&t=ff6081f1e5a590b3033596a43d165f3e",
		Salt:          pass.GetSalt(),
		Gender:        vobj.GenderMale,
		Role:          vobj.RoleSuperAdmin,
		Status:        vobj.StatusEnable,
	}

	return query.Use(d.mainDB).SysUser.Clauses(clause.OnConflict{DoNothing: true}).Create(user)
}

// syncBizDatabase 同步业务模型到各个团队， 保证数据一致性
func syncBizDatabase(d *Data) error {
	// 获取所有团队
	teams, err := query.Use(d.mainDB).SysTeam.Find()
	if err != nil {
		return err
	}
	for _, team := range teams {
		// 获取团队业务库连接
		db, err := d.GetBizGormDB(team.ID)
		if err != nil {
			return err
		}
		if err = db.AutoMigrate(bizmodel.Models()...); err != nil {
			return err
		}
		// TODO 同步实时告警数据库
		alarmDB, err := d.GetAlarmGormDB(team.ID)
		if err != nil {
			return err
		}
		if err = alarmDB.AutoMigrate(bizmodel.AlarmModels()...); err != nil {
			return err
		}
	}
	return nil
}
