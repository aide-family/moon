package promdict

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.PromDictRepo = (*promDictRepoImpl)(nil)

type promDictRepoImpl struct {
	repository.UnimplementedPromDictRepo
	data *data.Data
	log  *log.Helper
}

func (l *promDictRepoImpl) GetDictByIds(ctx context.Context, ids ...uint32) ([]*bo.DictBO, error) {
	dictList := make([]*do.SysDict, 0, len(ids))
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Find(&dictList).Error; err != nil {
		return nil, err
	}
	return slices.To(dictList, func(item *do.SysDict) *bo.DictBO { return bo.DictModelToBO(item) }), nil
}

func (l *promDictRepoImpl) CreateDict(ctx context.Context, dictBO *bo.DictBO) (*bo.DictBO, error) {
	newModelData := dictBO.ToModel()
	if err := l.data.DB().WithContext(ctx).Create(newModelData).Error; err != nil {
		return nil, err
	}

	return bo.DictModelToBO(newModelData), nil
}

func (l *promDictRepoImpl) UpdateDictById(ctx context.Context, id uint32, dictBO *bo.DictBO) (*bo.DictBO, error) {
	newModelData := dictBO.ToModel()
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(id)).Updates(newModelData).Error; err != nil {
		return nil, err
	}

	return bo.DictModelToBO(newModelData), nil
}

func (l *promDictRepoImpl) BatchUpdateDictStatusByIds(ctx context.Context, status vo.Status, ids []uint32) error {
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Updates(&do.SysDict{Status: status}).Error; err != nil {
		return err
	}
	return nil
}

func (l *promDictRepoImpl) DeleteDictByIds(ctx context.Context, id ...uint32) error {
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(id...)).Delete(&do.SysDict{}).Error; err != nil {
		return err
	}
	return nil
}

func (l *promDictRepoImpl) GetDictById(ctx context.Context, id uint32) (*bo.DictBO, error) {
	var detailModel do.SysDict
	if err := l.data.DB().WithContext(ctx).First(&detailModel, id).Error; err != nil {
		return nil, err
	}
	return bo.DictModelToBO(&detailModel), nil
}

func (l *promDictRepoImpl) ListDict(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.DictBO, error) {
	var dictModelList []*do.SysDict
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, bo.Page(pgInfo))...).Find(&dictModelList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.SysDict{}).Scopes(scopes...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	boList := slices.To(dictModelList, func(item *do.SysDict) *bo.DictBO {
		return bo.DictModelToBO(item)
	})

	return boList, nil
}

func NewPromDictRepo(data *data.Data, logger log.Logger) repository.PromDictRepo {
	return &promDictRepoImpl{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data.repository.dict")),
	}
}
