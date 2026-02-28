package impl

import (
	"context"

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
	"github.com/aide-family/marksman/internal/data/impl/do"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewStrategyGroupRepository(d *data.Data) (repository.StrategyGroup, error) {
	query.SetDefault(d.DB())
	return &strategyGroupRepository{db: d.DB()}, nil
}

type strategyGroupRepository struct {
	db *gorm.DB
}

func (r *strategyGroupRepository) CreateStrategyGroup(ctx context.Context, req *bo.CreateStrategyGroupBo) error {
	m := convert.ToStrategyGroupDo(ctx, req)
	return query.StrategyGroup.WithContext(ctx).Create(m)
}

func (r *strategyGroupRepository) UpdateStrategyGroup(ctx context.Context, req *bo.UpdateStrategyGroupBo) error {
	sg := query.StrategyGroup
	columns := []field.AssignExpr{
		sg.Name.Value(req.Name),
		sg.Remark.Value(req.Remark),
		sg.Metadata.Value(safety.NewMap(req.Metadata)),
	}
	_, err := sg.WithContext(ctx).Where(
		sg.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		sg.ID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(columns...)
	return err
}

func (r *strategyGroupRepository) UpdateStrategyGroupStatus(ctx context.Context, req *bo.UpdateStrategyGroupStatusBo) error {
	sg := query.StrategyGroup
	info, err := query.StrategyGroup.WithContext(ctx).Where(
		sg.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		sg.ID.Eq(req.UID.Int64()),
	).Update(sg.Status, req.Status)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy group not found")
	}
	return nil
}

func (r *strategyGroupRepository) DeleteStrategyGroup(ctx context.Context, uid snowflake.ID) error {
	sg := query.StrategyGroup
	info, err := query.StrategyGroup.WithContext(ctx).Where(
		sg.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		sg.ID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy group not found")
	}
	return nil
}

func (r *strategyGroupRepository) GetStrategyGroup(ctx context.Context, uid snowflake.ID) (*bo.StrategyGroupItemBo, error) {
	sg := query.StrategyGroup
	m, err := query.StrategyGroup.WithContext(ctx).Where(
		sg.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		sg.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, merr.ErrorNotFound("strategy group not found")
		}
		return nil, err
	}
	return convert.ToStrategyGroupItemBo(m), nil
}

func (r *strategyGroupRepository) ListStrategyGroup(ctx context.Context, req *bo.ListStrategyGroupBo) (*bo.PageResponseBo[*bo.StrategyGroupItemBo], error) {
	sg := query.StrategyGroup
	wrappers := sg.WithContext(ctx)
	wrappers = wrappers.Where(sg.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		wrappers = wrappers.Where(sg.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(sg.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	list, err := wrappers.Order(sg.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.StrategyGroupItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToStrategyGroupItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *strategyGroupRepository) SelectStrategyGroup(ctx context.Context, req *bo.SelectStrategyGroupBo) (*bo.SelectStrategyGroupBoResult, error) {
	sg := query.StrategyGroup
	wrappers := sg.WithContext(ctx)
	wrappers = wrappers.Where(sg.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		wrappers = wrappers.Where(sg.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(sg.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	if req.LastUID.Int64() > 0 {
		wrappers = wrappers.Where(sg.ID.Gt(req.LastUID.Int64()))
	}
	if req.Limit > 0 {
		wrappers = wrappers.Limit(int(req.Limit))
	}
	list, err := wrappers.Find()
	if err != nil {
		return nil, err
	}
	selectItems := make([]*bo.StrategyGroupItemSelectBo, 0, len(list))
	for _, m := range list {
		selectItems = append(selectItems, convert.ToStrategyGroupItemSelectBo(m))
	}
	var lastUID snowflake.ID
	if len(list) > 0 {
		lastUID = list[len(list)-1].ID
	}
	return &bo.SelectStrategyGroupBoResult{
		Items:   selectItems,
		Total:   total,
		LastUID: lastUID,
		HasMore: req.Limit > 0 && len(list) >= int(req.Limit),
	}, nil
}

func (r *strategyGroupRepository) StrategyGroupBindReceivers(ctx context.Context, req *bo.StrategyGroupBindReceiversBo) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		_, err := q.StrategyGroupReceiver.WithContext(ctx).Where(
			q.StrategyGroupReceiver.StrategyGroupUID.Eq(req.StrategyGroupUID.Int64()),
		).Delete()
		if err != nil {
			return err
		}
		if len(req.ReceiverUIDs) == 0 {
			return nil
		}
		rows := make([]*do.StrategyGroupReceiver, 0, len(req.ReceiverUIDs))
		for _, recUID := range req.ReceiverUIDs {
			rows = append(rows, &do.StrategyGroupReceiver{
				StrategyGroupUID: req.StrategyGroupUID,
				ReceiverUID:      recUID,
			})
		}
		return tx.CreateInBatches(rows, 100).Error
	})
}
