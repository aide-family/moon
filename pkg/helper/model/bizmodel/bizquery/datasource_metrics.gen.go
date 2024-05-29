// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package bizquery

import (
	"context"

	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"
)

func newDatasourceMetric(db *gorm.DB, opts ...gen.DOOption) datasourceMetric {
	_datasourceMetric := datasourceMetric{}

	_datasourceMetric.datasourceMetricDo.UseDB(db, opts...)
	_datasourceMetric.datasourceMetricDo.UseModel(&bizmodel.DatasourceMetric{})

	tableName := _datasourceMetric.datasourceMetricDo.TableName()
	_datasourceMetric.ALL = field.NewAsterisk(tableName)
	_datasourceMetric.ID = field.NewUint32(tableName, "id")
	_datasourceMetric.Name = field.NewString(tableName, "name")
	_datasourceMetric.Category = field.NewInt(tableName, "category")
	_datasourceMetric.Unit = field.NewString(tableName, "unit")
	_datasourceMetric.Remark = field.NewString(tableName, "remark")
	_datasourceMetric.DatasourceID = field.NewUint32(tableName, "datasource_id")
	_datasourceMetric.CreatedAt = field.NewField(tableName, "created_at")
	_datasourceMetric.UpdatedAt = field.NewField(tableName, "updated_at")
	_datasourceMetric.DeletedAt = field.NewInt64(tableName, "deleted_at")
	_datasourceMetric.Labels = datasourceMetricHasManyLabels{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Labels", "bizmodel.MetricLabel"),
		LabelValues: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Labels.LabelValues", "bizmodel.MetricLabelValue"),
		},
	}

	_datasourceMetric.fillFieldMap()

	return _datasourceMetric
}

type datasourceMetric struct {
	datasourceMetricDo

	ALL          field.Asterisk
	ID           field.Uint32
	Name         field.String
	Category     field.Int
	Unit         field.String
	Remark       field.String
	DatasourceID field.Uint32
	CreatedAt    field.Field
	UpdatedAt    field.Field
	DeletedAt    field.Int64
	Labels       datasourceMetricHasManyLabels

	fieldMap map[string]field.Expr
}

func (d datasourceMetric) Table(newTableName string) *datasourceMetric {
	d.datasourceMetricDo.UseTable(newTableName)
	return d.updateTableName(newTableName)
}

func (d datasourceMetric) As(alias string) *datasourceMetric {
	d.datasourceMetricDo.DO = *(d.datasourceMetricDo.As(alias).(*gen.DO))
	return d.updateTableName(alias)
}

func (d *datasourceMetric) updateTableName(table string) *datasourceMetric {
	d.ALL = field.NewAsterisk(table)
	d.ID = field.NewUint32(table, "id")
	d.Name = field.NewString(table, "name")
	d.Category = field.NewInt(table, "category")
	d.Unit = field.NewString(table, "unit")
	d.Remark = field.NewString(table, "remark")
	d.DatasourceID = field.NewUint32(table, "datasource_id")
	d.CreatedAt = field.NewField(table, "created_at")
	d.UpdatedAt = field.NewField(table, "updated_at")
	d.DeletedAt = field.NewInt64(table, "deleted_at")

	d.fillFieldMap()

	return d
}

func (d *datasourceMetric) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := d.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (d *datasourceMetric) fillFieldMap() {
	d.fieldMap = make(map[string]field.Expr, 10)
	d.fieldMap["id"] = d.ID
	d.fieldMap["name"] = d.Name
	d.fieldMap["category"] = d.Category
	d.fieldMap["unit"] = d.Unit
	d.fieldMap["remark"] = d.Remark
	d.fieldMap["datasource_id"] = d.DatasourceID
	d.fieldMap["created_at"] = d.CreatedAt
	d.fieldMap["updated_at"] = d.UpdatedAt
	d.fieldMap["deleted_at"] = d.DeletedAt

}

func (d datasourceMetric) clone(db *gorm.DB) datasourceMetric {
	d.datasourceMetricDo.ReplaceConnPool(db.Statement.ConnPool)
	return d
}

func (d datasourceMetric) replaceDB(db *gorm.DB) datasourceMetric {
	d.datasourceMetricDo.ReplaceDB(db)
	return d
}

type datasourceMetricHasManyLabels struct {
	db *gorm.DB

	field.RelationField

	LabelValues struct {
		field.RelationField
	}
}

func (a datasourceMetricHasManyLabels) Where(conds ...field.Expr) *datasourceMetricHasManyLabels {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a datasourceMetricHasManyLabels) WithContext(ctx context.Context) *datasourceMetricHasManyLabels {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a datasourceMetricHasManyLabels) Session(session *gorm.Session) *datasourceMetricHasManyLabels {
	a.db = a.db.Session(session)
	return &a
}

func (a datasourceMetricHasManyLabels) Model(m *bizmodel.DatasourceMetric) *datasourceMetricHasManyLabelsTx {
	return &datasourceMetricHasManyLabelsTx{a.db.Model(m).Association(a.Name())}
}

type datasourceMetricHasManyLabelsTx struct{ tx *gorm.Association }

func (a datasourceMetricHasManyLabelsTx) Find() (result []*bizmodel.MetricLabel, err error) {
	return result, a.tx.Find(&result)
}

func (a datasourceMetricHasManyLabelsTx) Append(values ...*bizmodel.MetricLabel) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a datasourceMetricHasManyLabelsTx) Replace(values ...*bizmodel.MetricLabel) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a datasourceMetricHasManyLabelsTx) Delete(values ...*bizmodel.MetricLabel) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a datasourceMetricHasManyLabelsTx) Clear() error {
	return a.tx.Clear()
}

func (a datasourceMetricHasManyLabelsTx) Count() int64 {
	return a.tx.Count()
}

type datasourceMetricDo struct{ gen.DO }

type IDatasourceMetricDo interface {
	gen.SubQuery
	Debug() IDatasourceMetricDo
	WithContext(ctx context.Context) IDatasourceMetricDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IDatasourceMetricDo
	WriteDB() IDatasourceMetricDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IDatasourceMetricDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IDatasourceMetricDo
	Not(conds ...gen.Condition) IDatasourceMetricDo
	Or(conds ...gen.Condition) IDatasourceMetricDo
	Select(conds ...field.Expr) IDatasourceMetricDo
	Where(conds ...gen.Condition) IDatasourceMetricDo
	Order(conds ...field.Expr) IDatasourceMetricDo
	Distinct(cols ...field.Expr) IDatasourceMetricDo
	Omit(cols ...field.Expr) IDatasourceMetricDo
	Join(table schema.Tabler, on ...field.Expr) IDatasourceMetricDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IDatasourceMetricDo
	RightJoin(table schema.Tabler, on ...field.Expr) IDatasourceMetricDo
	Group(cols ...field.Expr) IDatasourceMetricDo
	Having(conds ...gen.Condition) IDatasourceMetricDo
	Limit(limit int) IDatasourceMetricDo
	Offset(offset int) IDatasourceMetricDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IDatasourceMetricDo
	Unscoped() IDatasourceMetricDo
	Create(values ...*bizmodel.DatasourceMetric) error
	CreateInBatches(values []*bizmodel.DatasourceMetric, batchSize int) error
	Save(values ...*bizmodel.DatasourceMetric) error
	First() (*bizmodel.DatasourceMetric, error)
	Take() (*bizmodel.DatasourceMetric, error)
	Last() (*bizmodel.DatasourceMetric, error)
	Find() ([]*bizmodel.DatasourceMetric, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*bizmodel.DatasourceMetric, err error)
	FindInBatches(result *[]*bizmodel.DatasourceMetric, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*bizmodel.DatasourceMetric) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IDatasourceMetricDo
	Assign(attrs ...field.AssignExpr) IDatasourceMetricDo
	Joins(fields ...field.RelationField) IDatasourceMetricDo
	Preload(fields ...field.RelationField) IDatasourceMetricDo
	FirstOrInit() (*bizmodel.DatasourceMetric, error)
	FirstOrCreate() (*bizmodel.DatasourceMetric, error)
	FindByPage(offset int, limit int) (result []*bizmodel.DatasourceMetric, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IDatasourceMetricDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (d datasourceMetricDo) Debug() IDatasourceMetricDo {
	return d.withDO(d.DO.Debug())
}

func (d datasourceMetricDo) WithContext(ctx context.Context) IDatasourceMetricDo {
	return d.withDO(d.DO.WithContext(ctx))
}

func (d datasourceMetricDo) ReadDB() IDatasourceMetricDo {
	return d.Clauses(dbresolver.Read)
}

func (d datasourceMetricDo) WriteDB() IDatasourceMetricDo {
	return d.Clauses(dbresolver.Write)
}

func (d datasourceMetricDo) Session(config *gorm.Session) IDatasourceMetricDo {
	return d.withDO(d.DO.Session(config))
}

func (d datasourceMetricDo) Clauses(conds ...clause.Expression) IDatasourceMetricDo {
	return d.withDO(d.DO.Clauses(conds...))
}

func (d datasourceMetricDo) Returning(value interface{}, columns ...string) IDatasourceMetricDo {
	return d.withDO(d.DO.Returning(value, columns...))
}

func (d datasourceMetricDo) Not(conds ...gen.Condition) IDatasourceMetricDo {
	return d.withDO(d.DO.Not(conds...))
}

func (d datasourceMetricDo) Or(conds ...gen.Condition) IDatasourceMetricDo {
	return d.withDO(d.DO.Or(conds...))
}

func (d datasourceMetricDo) Select(conds ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.Select(conds...))
}

func (d datasourceMetricDo) Where(conds ...gen.Condition) IDatasourceMetricDo {
	return d.withDO(d.DO.Where(conds...))
}

func (d datasourceMetricDo) Order(conds ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.Order(conds...))
}

func (d datasourceMetricDo) Distinct(cols ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.Distinct(cols...))
}

func (d datasourceMetricDo) Omit(cols ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.Omit(cols...))
}

func (d datasourceMetricDo) Join(table schema.Tabler, on ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.Join(table, on...))
}

func (d datasourceMetricDo) LeftJoin(table schema.Tabler, on ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.LeftJoin(table, on...))
}

func (d datasourceMetricDo) RightJoin(table schema.Tabler, on ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.RightJoin(table, on...))
}

func (d datasourceMetricDo) Group(cols ...field.Expr) IDatasourceMetricDo {
	return d.withDO(d.DO.Group(cols...))
}

func (d datasourceMetricDo) Having(conds ...gen.Condition) IDatasourceMetricDo {
	return d.withDO(d.DO.Having(conds...))
}

func (d datasourceMetricDo) Limit(limit int) IDatasourceMetricDo {
	return d.withDO(d.DO.Limit(limit))
}

func (d datasourceMetricDo) Offset(offset int) IDatasourceMetricDo {
	return d.withDO(d.DO.Offset(offset))
}

func (d datasourceMetricDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IDatasourceMetricDo {
	return d.withDO(d.DO.Scopes(funcs...))
}

func (d datasourceMetricDo) Unscoped() IDatasourceMetricDo {
	return d.withDO(d.DO.Unscoped())
}

func (d datasourceMetricDo) Create(values ...*bizmodel.DatasourceMetric) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Create(values)
}

func (d datasourceMetricDo) CreateInBatches(values []*bizmodel.DatasourceMetric, batchSize int) error {
	return d.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (d datasourceMetricDo) Save(values ...*bizmodel.DatasourceMetric) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Save(values)
}

func (d datasourceMetricDo) First() (*bizmodel.DatasourceMetric, error) {
	if result, err := d.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*bizmodel.DatasourceMetric), nil
	}
}

func (d datasourceMetricDo) Take() (*bizmodel.DatasourceMetric, error) {
	if result, err := d.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*bizmodel.DatasourceMetric), nil
	}
}

func (d datasourceMetricDo) Last() (*bizmodel.DatasourceMetric, error) {
	if result, err := d.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*bizmodel.DatasourceMetric), nil
	}
}

func (d datasourceMetricDo) Find() ([]*bizmodel.DatasourceMetric, error) {
	result, err := d.DO.Find()
	return result.([]*bizmodel.DatasourceMetric), err
}

func (d datasourceMetricDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*bizmodel.DatasourceMetric, err error) {
	buf := make([]*bizmodel.DatasourceMetric, 0, batchSize)
	err = d.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (d datasourceMetricDo) FindInBatches(result *[]*bizmodel.DatasourceMetric, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return d.DO.FindInBatches(result, batchSize, fc)
}

func (d datasourceMetricDo) Attrs(attrs ...field.AssignExpr) IDatasourceMetricDo {
	return d.withDO(d.DO.Attrs(attrs...))
}

func (d datasourceMetricDo) Assign(attrs ...field.AssignExpr) IDatasourceMetricDo {
	return d.withDO(d.DO.Assign(attrs...))
}

func (d datasourceMetricDo) Joins(fields ...field.RelationField) IDatasourceMetricDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Joins(_f))
	}
	return &d
}

func (d datasourceMetricDo) Preload(fields ...field.RelationField) IDatasourceMetricDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Preload(_f))
	}
	return &d
}

func (d datasourceMetricDo) FirstOrInit() (*bizmodel.DatasourceMetric, error) {
	if result, err := d.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*bizmodel.DatasourceMetric), nil
	}
}

func (d datasourceMetricDo) FirstOrCreate() (*bizmodel.DatasourceMetric, error) {
	if result, err := d.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*bizmodel.DatasourceMetric), nil
	}
}

func (d datasourceMetricDo) FindByPage(offset int, limit int) (result []*bizmodel.DatasourceMetric, count int64, err error) {
	result, err = d.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = d.Offset(-1).Limit(-1).Count()
	return
}

func (d datasourceMetricDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	err = d.Offset(offset).Limit(limit).Scan(result)
	return
}

func (d datasourceMetricDo) Scan(result interface{}) (err error) {
	return d.DO.Scan(result)
}

func (d datasourceMetricDo) Delete(models ...*bizmodel.DatasourceMetric) (result gen.ResultInfo, err error) {
	return d.DO.Delete(models)
}

func (d *datasourceMetricDo) withDO(do gen.Dao) *datasourceMetricDo {
	d.DO = *do.(*gen.DO)
	return d
}
