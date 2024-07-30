package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/imodel"
)

type (
	// Dict 字典接口
	Dict interface {
		// Create 创建字典
		Create(context.Context, *bo.CreateDictParams) (imodel.IDict, error)

		// GetByID 通过id 获取字典详情
		GetByID(context.Context, uint32) (imodel.IDict, error)

		// FindByPage 分页查询字典列表
		FindByPage(context.Context, *bo.QueryDictListParams) ([]imodel.IDict, error)

		// DeleteByID 通过ID删除字典
		DeleteByID(context.Context, uint32) error

		// UpdateStatusByIds 通过ID列表批量更新字典状态
		UpdateStatusByIds(context.Context, *bo.UpdateDictStatusParams) error

		// UpdateByID 通过ID更新字典数据
		UpdateByID(context.Context, *bo.UpdateDictParams) error
	}
)
