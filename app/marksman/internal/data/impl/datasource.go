package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewDatasourceRepository(d *data.Data) (repository.Datasource, error) {
	query.SetDefault(d.DB())
	return &datasourceRepository{db: d.DB()}, nil
}

type datasourceRepository struct {
	db *gorm.DB
}

func (r *datasourceRepository) CheckDatasourceNameExist(ctx context.Context, name string, uid ...snowflake.ID) error {
	d := query.Datasource
	wrappers := d.WithContext(ctx)
	wrappers = wrappers.Where(d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if len(uid) > 0 {
		wrappers = wrappers.Where(d.ID.Neq(uid[0].Int64()))
	}
	datasourceDo, err := wrappers.Where(d.Name.Eq(name)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if len(uid) > 0 && datasourceDo.ID == uid[0] {
		return nil
	}
	return merr.ErrorParams("datasource name already exists, uid: %d", datasourceDo.ID.Int64())
}

func (r *datasourceRepository) CreateDatasource(ctx context.Context, req *bo.CreateDatasourceBo) (snowflake.ID, error) {
	m := convert.ToDatasourceDo(ctx, req)
	if err := query.Datasource.WithContext(ctx).Create(m); err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (r *datasourceRepository) UpdateDatasource(ctx context.Context, req *bo.UpdateDatasourceBo) error {
	d := query.Datasource
	columns := []field.AssignExpr{
		d.Name.Value(req.Name),
		d.Type.Value(int32(req.Type)),
		d.Driver.Value(int32(req.Driver)),
		d.Metadata.Value(safety.NewMap(req.Metadata)),
		d.URL.Value(req.URL),
		d.Remark.Value(req.Remark),
	}
	_, err := d.WithContext(ctx).Where(
		d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		d.ID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(columns...)
	return err
}

func (r *datasourceRepository) DeleteDatasource(ctx context.Context, uid snowflake.ID) error {
	d := query.Datasource
	info, err := query.Datasource.WithContext(ctx).Where(
		d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		d.ID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("datasource not found")
	}
	return nil
}

func (r *datasourceRepository) GetDatasource(ctx context.Context, uid snowflake.ID) (*bo.DatasourceItemBo, error) {
	d := query.Datasource
	m, err := query.Datasource.WithContext(ctx).Where(
		d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		d.ID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("datasource not found")
		}
		return nil, err
	}
	return convert.ToDatasourceItemBo(m), nil
}

func (r *datasourceRepository) ListDatasource(ctx context.Context, req *bo.ListDatasourceBo) (*bo.PageResponseBo[*bo.DatasourceItemBo], error) {
	d := query.Datasource
	wrappers := d.WithContext(ctx)
	wrappers = wrappers.Where(d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		k := "%" + strings.TrimSpace(req.Keyword) + "%"
		wrappers = wrappers.Or(d.Name.Like(k), d.URL.Like(k), d.Remark.Like(k))
	}
	if req.Type != enum.DatasourceType_DatasourceType_UNKNOWN {
		wrappers = wrappers.Where(d.Type.Eq(int32(req.Type)))
	}
	if req.Driver != enum.DatasourceDriver_DatasourceDriver_UNKNOWN {
		wrappers = wrappers.Where(d.Driver.Eq(int32(req.Driver)))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(d.Status.Eq(int32(req.Status)))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	if req.Page > 0 && req.PageSize > 0 {
		wrappers = wrappers.Offset(req.Offset()).Limit(req.Limit())
	}
	list, err := wrappers.Order(d.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.DatasourceItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToDatasourceItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *datasourceRepository) SelectDatasource(ctx context.Context, req *bo.SelectDatasourceBo) (*bo.SelectDatasourceReplyBo, error) {
	d := query.Datasource
	wrappers := d.WithContext(ctx)
	wrappers = wrappers.Where(d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if req.Keyword != "" {
		k := "%" + strings.TrimSpace(req.Keyword) + "%"
		wrappers = wrappers.Or(d.Name.Like(k), d.URL.Like(k), d.Remark.Like(k))
	}
	if req.Type != enum.DatasourceType_DatasourceType_UNKNOWN {
		wrappers = wrappers.Where(d.Type.Eq(int32(req.Type)))
	}
	if req.Driver != enum.DatasourceDriver_DatasourceDriver_UNKNOWN {
		wrappers = wrappers.Where(d.Driver.Eq(int32(req.Driver)))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(d.Status.Eq(int32(req.Status)))
	}
	if len(req.Uids) > 0 {
		wrappers = wrappers.Where(d.ID.In(req.Uids...))
	}
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}

	if req.LastUID > 0 {
		wrappers = wrappers.Where(d.ID.Gt(req.LastUID.Int64()))
	}
	wrappers = wrappers.Limit(int(req.Limit))
	list, err := wrappers.Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.SelectDatasourceItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToSelectDatasourceItemBo(m))
	}
	return &bo.SelectDatasourceReplyBo{
		Items:   items,
		Total:   total,
		LastUID: req.LastUID,
		HasMore: len(list) >= int(req.Limit),
	}, nil
}

func (r *datasourceRepository) ListAllForProbe(ctx context.Context, batchSize int) ([]*bo.DatasourceItemBo, error) {
	if batchSize <= 0 {
		batchSize = 200
	}
	d := query.Datasource
	var all []*bo.DatasourceItemBo
	for offset := 0; ; offset += batchSize {
		list, err := d.WithContext(ctx).
			Where(d.Status.Eq(int32(enum.GlobalStatus_ENABLED))).
			Order(d.ID.Desc()).
			Offset(offset).
			Limit(batchSize).
			Find()
		if err != nil {
			return nil, err
		}
		for _, m := range list {
			all = append(all, convert.ToDatasourceItemBo(m))
		}
		if len(list) < batchSize {
			break
		}
	}
	return all, nil
}
