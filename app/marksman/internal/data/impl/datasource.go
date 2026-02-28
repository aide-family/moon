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

func NewDatasourceRepository(d *data.Data) (repository.Datasource, error) {
	query.SetDefault(d.DB())
	return &datasourceRepository{db: d.DB()}, nil
}

type datasourceRepository struct {
	db *gorm.DB
}

func (r *datasourceRepository) CreateDatasource(ctx context.Context, req *bo.CreateDatasourceBo) error {
	m := convert.ToDatasourceDo(req)
	return query.Datasource.WithContext(ctx).Create(m)
}

func (r *datasourceRepository) UpdateDatasource(ctx context.Context, req *bo.UpdateDatasourceBo) error {
	d := query.Datasource
	columns := []field.AssignExpr{
		d.Name.Value(req.Name),
		d.Type.Value(int32(req.Type)),
		d.Driver.Value(int32(req.Driver)),
		d.Metadata.Value(safety.NewMap(req.Metadata)),
	}
	_, err := d.WithContext(ctx).Where(
		d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		d.UID.Eq(req.UID.Int64()),
	).UpdateColumnSimple(columns...)
	return err
}

func (r *datasourceRepository) DeleteDatasource(ctx context.Context, uid snowflake.ID) error {
	d := query.Datasource
	info, err := query.Datasource.WithContext(ctx).Where(
		d.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()),
		d.UID.Eq(uid.Int64()),
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
		d.UID.Eq(uid.Int64()),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
		wrappers = wrappers.Where(d.Name.Like(k))
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
	list, err := wrappers.Order(d.UID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.DatasourceItemBo, 0, len(list))
	for _, m := range list {
		items = append(items, convert.ToDatasourceItemBo(m))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}
