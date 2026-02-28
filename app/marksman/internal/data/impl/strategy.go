package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewStrategyRepository(d *data.Data) (repository.Strategy, error) {
	query.SetDefault(d.DB())
	return &strategyRepository{db: d.DB()}, nil
}

type strategyRepository struct {
	db *gorm.DB
}

func (r *strategyRepository) CreateStrategy(ctx context.Context, req *bo.CreateStrategyBo) error {
	m := convert.ToStrategyDo(req)
	return query.Strategy.WithContext(ctx).Create(m)
}

func (r *strategyRepository) UpdateStrategy(ctx context.Context, req *bo.UpdateStrategyBo) error {
	s := query.Strategy
	columns := []field.AssignExpr{
		s.Name.Value(req.Name),
		s.Remark.Value(req.Remark),
		s.StrategyGroupUID.Value(req.StrategyGroupUID.Int64()),
		s.Type.Value(int32(req.Type)),
		s.Driver.Value(int32(req.Driver)),
		s.Metadata.Value(safety.NewMap(req.Metadata)),
	}
	_, err := s.WithContext(ctx).Where(
		s.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		s.UID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(columns...)
	return err
}

func (r *strategyRepository) UpdateStrategyStatus(ctx context.Context, req *bo.UpdateStrategyStatusBo) error {
	s := query.Strategy
	info, err := query.Strategy.WithContext(ctx).Where(
		s.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		s.UID.Eq(req.UID.Int64()),
	).Update(s.Status, req.Status)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy not found")
	}
	return nil
}

func (r *strategyRepository) DeleteStrategy(ctx context.Context, uid snowflake.ID) error {
	s := query.Strategy
	info, err := query.Strategy.WithContext(ctx).Where(
		s.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		s.UID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy not found")
	}
	return nil
}

func (r *strategyRepository) GetStrategy(ctx context.Context, uid snowflake.ID) (*bo.StrategyItemBo, error) {
	s := query.Strategy
	m, err := query.Strategy.WithContext(ctx).Where(
		s.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		s.UID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, merr.ErrorNotFound("strategy not found")
		}
		return nil, err
	}
	return convert.ToStrategyItemBo(m), nil
}

func (r *strategyRepository) ListStrategy(ctx context.Context, req *bo.ListStrategyBo) (*bo.PageResponseBo[*bo.StrategyItemBo], error) {
	s := query.Strategy
	wrappers := s.WithContext(ctx)
	wrappers = wrappers.Where(s.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.StrategyGroupUID.Int64() > 0 {
		wrappers = wrappers.Where(s.StrategyGroupUID.Eq(req.StrategyGroupUID.Int64()))
	}
	if req.Keyword != "" {
		k := "%" + strings.TrimSpace(req.Keyword) + "%"
		wrappers = wrappers.Where(s.Name.Like(k))
	}
	if req.Type != enum.DatasourceType_DatasourceType_UNKNOWN {
		wrappers = wrappers.Where(s.Type.Eq(int32(req.Type)))
	}
	if req.Driver != enum.DatasourceDriver_DatasourceDriver_UNKNOWN {
		wrappers = wrappers.Where(s.Driver.Eq(int32(req.Driver)))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(s.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	list, err := wrappers.Order(s.UID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.StrategyItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToStrategyItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}
