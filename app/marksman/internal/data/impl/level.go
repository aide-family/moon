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
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewLevelRepository(d *data.Data) (repository.Level, error) {
	query.SetDefault(d.DB())
	return &levelRepository{db: d.DB()}, nil
}

type levelRepository struct {
	db *gorm.DB
}

func (r *levelRepository) CreateLevel(ctx context.Context, req *bo.CreateLevelBo) error {
	m := convert.ToLevelDo(req)
	return query.Level.WithContext(ctx).Create(m)
}

func (r *levelRepository) UpdateLevel(ctx context.Context, req *bo.UpdateLevelBo) error {
	l := query.Level
	columns := []field.AssignExpr{
		l.Name.Value(req.Name),
		l.Remark.Value(req.Remark),
		l.Metadata.Value(safety.NewMap(req.Metadata)),
	}
	_, err := l.WithContext(ctx).Where(l.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), l.UID.Eq(req.UID.Int64())).UpdateColumnSimple(columns...)
	return err
}

func (r *levelRepository) UpdateLevelStatus(ctx context.Context, req *bo.UpdateLevelStatusBo) error {
	l := query.Level
	info, err := query.Level.WithContext(ctx).Where(
		l.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		l.UID.Eq(req.UID.Int64()),
	).Update(l.Status, req.Status)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("level not found")
	}
	return nil
}

func (r *levelRepository) DeleteLevel(ctx context.Context, uid snowflake.ID) error {
	l := query.Level
	info, err := query.Level.WithContext(ctx).Where(
		l.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		l.UID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("level not found")
	}
	return nil
}

func (r *levelRepository) GetLevel(ctx context.Context, uid snowflake.ID) (*bo.LevelItemBo, error) {
	l := query.Level
	m, err := query.Level.WithContext(ctx).Where(
		l.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		l.UID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, merr.ErrorNotFound("level not found")
		}
		return nil, err
	}
	return convert.ToLevelItemBo(m), nil
}

func (r *levelRepository) ListLevel(ctx context.Context, req *bo.ListLevelBo) (*bo.PageResponseBo[*bo.LevelItemBo], error) {
	l := query.Level
	wrappers := l.WithContext(ctx)
	wrappers = wrappers.Where(l.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		wrappers = wrappers.Where(l.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(l.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	list, err := wrappers.Find()
	if err != nil {
		return nil, err
	}
	levelItems := make([]*bo.LevelItemBo, 0, len(list))
	for _, m := range list {
		levelItems = append(levelItems, convert.ToLevelItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, levelItems), nil
}

func (r *levelRepository) SelectLevel(ctx context.Context, req *bo.SelectLevelBo) (*bo.SelectLevelBoResult, error) {
	l := query.Level
	wrappers := l.WithContext(ctx)
	wrappers = wrappers.Where(l.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		wrappers = wrappers.Where(l.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(l.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	if req.LastUID > 0 {
		wrappers = wrappers.Where(l.UID.Gt(req.LastUID.Int64()))
	}
	wrappers = wrappers.Limit(int(req.Limit))
	list, err := wrappers.Find()
	if err != nil {
		return nil, err
	}
	levelItems := make([]*bo.LevelItemSelectBo, 0, len(list))
	for _, m := range list {
		levelItems = append(levelItems, convert.ToLevelItemSelectBo(m))
	}
	var lastUID snowflake.ID
	if len(list) > 0 {
		lastUID = list[len(list)-1].UID
	}
	return &bo.SelectLevelBoResult{
		Items:   levelItems,
		Total:   total,
		LastUID: lastUID,
		HasMore: len(list) >= int(req.Limit),
	}, nil
}
