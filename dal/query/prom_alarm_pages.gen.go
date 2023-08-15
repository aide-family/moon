// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"prometheus-manager/dal/model"
)

func newPromAlarmPage(db *gorm.DB, opts ...gen.DOOption) promAlarmPage {
	_promAlarmPage := promAlarmPage{}

	_promAlarmPage.promAlarmPageDo.UseDB(db, opts...)
	_promAlarmPage.promAlarmPageDo.UseModel(&model.PromAlarmPage{})

	tableName := _promAlarmPage.promAlarmPageDo.TableName()
	_promAlarmPage.ALL = field.NewAsterisk(tableName)
	_promAlarmPage.ID = field.NewInt32(tableName, "id")
	_promAlarmPage.Name = field.NewString(tableName, "name")
	_promAlarmPage.Remark = field.NewString(tableName, "remark")
	_promAlarmPage.Icon = field.NewString(tableName, "icon")
	_promAlarmPage.Color = field.NewString(tableName, "color")
	_promAlarmPage.Status = field.NewInt32(tableName, "status")
	_promAlarmPage.CreatedAt = field.NewTime(tableName, "created_at")
	_promAlarmPage.UpdatedAt = field.NewTime(tableName, "updated_at")
	_promAlarmPage.DeletedAt = field.NewField(tableName, "deleted_at")
	_promAlarmPage.PromStrategies = promAlarmPageHasManyPromStrategies{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("PromStrategies", "model.PromStrategy"),
	}

	_promAlarmPage.fillFieldMap()

	return _promAlarmPage
}

type promAlarmPage struct {
	promAlarmPageDo

	ALL            field.Asterisk
	ID             field.Int32
	Name           field.String // 报警页面名称
	Remark         field.String // 描述信息
	Icon           field.String // 图表
	Color          field.String // tab颜色
	Status         field.Int32  // 启用状态,1启用;2禁用
	CreatedAt      field.Time   // 创建时间
	UpdatedAt      field.Time   // 更新时间
	DeletedAt      field.Field  // 删除时间
	PromStrategies promAlarmPageHasManyPromStrategies

	fieldMap map[string]field.Expr
}

func (p promAlarmPage) Table(newTableName string) *promAlarmPage {
	p.promAlarmPageDo.UseTable(newTableName)
	return p.updateTableName(newTableName)
}

func (p promAlarmPage) As(alias string) *promAlarmPage {
	p.promAlarmPageDo.DO = *(p.promAlarmPageDo.As(alias).(*gen.DO))
	return p.updateTableName(alias)
}

func (p *promAlarmPage) updateTableName(table string) *promAlarmPage {
	p.ALL = field.NewAsterisk(table)
	p.ID = field.NewInt32(table, "id")
	p.Name = field.NewString(table, "name")
	p.Remark = field.NewString(table, "remark")
	p.Icon = field.NewString(table, "icon")
	p.Color = field.NewString(table, "color")
	p.Status = field.NewInt32(table, "status")
	p.CreatedAt = field.NewTime(table, "created_at")
	p.UpdatedAt = field.NewTime(table, "updated_at")
	p.DeletedAt = field.NewField(table, "deleted_at")

	p.fillFieldMap()

	return p
}

func (p *promAlarmPage) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := p.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (p *promAlarmPage) fillFieldMap() {
	p.fieldMap = make(map[string]field.Expr, 10)
	p.fieldMap["id"] = p.ID
	p.fieldMap["name"] = p.Name
	p.fieldMap["remark"] = p.Remark
	p.fieldMap["icon"] = p.Icon
	p.fieldMap["color"] = p.Color
	p.fieldMap["status"] = p.Status
	p.fieldMap["created_at"] = p.CreatedAt
	p.fieldMap["updated_at"] = p.UpdatedAt
	p.fieldMap["deleted_at"] = p.DeletedAt

}

func (p promAlarmPage) clone(db *gorm.DB) promAlarmPage {
	p.promAlarmPageDo.ReplaceConnPool(db.Statement.ConnPool)
	return p
}

func (p promAlarmPage) replaceDB(db *gorm.DB) promAlarmPage {
	p.promAlarmPageDo.ReplaceDB(db)
	return p
}

type promAlarmPageHasManyPromStrategies struct {
	db *gorm.DB

	field.RelationField
}

func (a promAlarmPageHasManyPromStrategies) Where(conds ...field.Expr) *promAlarmPageHasManyPromStrategies {
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

func (a promAlarmPageHasManyPromStrategies) WithContext(ctx context.Context) *promAlarmPageHasManyPromStrategies {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a promAlarmPageHasManyPromStrategies) Session(session *gorm.Session) *promAlarmPageHasManyPromStrategies {
	a.db = a.db.Session(session)
	return &a
}

func (a promAlarmPageHasManyPromStrategies) Model(m *model.PromAlarmPage) *promAlarmPageHasManyPromStrategiesTx {
	return &promAlarmPageHasManyPromStrategiesTx{a.db.Model(m).Association(a.Name())}
}

type promAlarmPageHasManyPromStrategiesTx struct{ tx *gorm.Association }

func (a promAlarmPageHasManyPromStrategiesTx) Find() (result []*model.PromStrategy, err error) {
	return result, a.tx.Find(&result)
}

func (a promAlarmPageHasManyPromStrategiesTx) Append(values ...*model.PromStrategy) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a promAlarmPageHasManyPromStrategiesTx) Replace(values ...*model.PromStrategy) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a promAlarmPageHasManyPromStrategiesTx) Delete(values ...*model.PromStrategy) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a promAlarmPageHasManyPromStrategiesTx) Clear() error {
	return a.tx.Clear()
}

func (a promAlarmPageHasManyPromStrategiesTx) Count() int64 {
	return a.tx.Count()
}

type promAlarmPageDo struct{ gen.DO }

type IPromAlarmPageDo interface {
	gen.SubQuery
	Debug() IPromAlarmPageDo
	WithContext(ctx context.Context) IPromAlarmPageDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IPromAlarmPageDo
	WriteDB() IPromAlarmPageDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IPromAlarmPageDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IPromAlarmPageDo
	Not(conds ...gen.Condition) IPromAlarmPageDo
	Or(conds ...gen.Condition) IPromAlarmPageDo
	Select(conds ...field.Expr) IPromAlarmPageDo
	Where(conds ...gen.Condition) IPromAlarmPageDo
	Order(conds ...field.Expr) IPromAlarmPageDo
	Distinct(cols ...field.Expr) IPromAlarmPageDo
	Omit(cols ...field.Expr) IPromAlarmPageDo
	Join(table schema.Tabler, on ...field.Expr) IPromAlarmPageDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IPromAlarmPageDo
	RightJoin(table schema.Tabler, on ...field.Expr) IPromAlarmPageDo
	Group(cols ...field.Expr) IPromAlarmPageDo
	Having(conds ...gen.Condition) IPromAlarmPageDo
	Limit(limit int) IPromAlarmPageDo
	Offset(offset int) IPromAlarmPageDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IPromAlarmPageDo
	Unscoped() IPromAlarmPageDo
	Create(values ...*model.PromAlarmPage) error
	CreateInBatches(values []*model.PromAlarmPage, batchSize int) error
	Save(values ...*model.PromAlarmPage) error
	First() (*model.PromAlarmPage, error)
	Take() (*model.PromAlarmPage, error)
	Last() (*model.PromAlarmPage, error)
	Find() ([]*model.PromAlarmPage, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.PromAlarmPage, err error)
	FindInBatches(result *[]*model.PromAlarmPage, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.PromAlarmPage) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IPromAlarmPageDo
	Assign(attrs ...field.AssignExpr) IPromAlarmPageDo
	Joins(fields ...field.RelationField) IPromAlarmPageDo
	Preload(fields ...field.RelationField) IPromAlarmPageDo
	FirstOrInit() (*model.PromAlarmPage, error)
	FirstOrCreate() (*model.PromAlarmPage, error)
	FindByPage(offset int, limit int) (result []*model.PromAlarmPage, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IPromAlarmPageDo
	UnderlyingDB() *gorm.DB
	schema.Tabler

	SaFindById(ctx context.Context, id int32) (result *model.PromAlarmPage, err error)
}

// select * from @@table where id = @id
func (p promAlarmPageDo) SaFindById(ctx context.Context, id int32) (result *model.PromAlarmPage, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("select * from prom_alarm_pages where id = ? ")

	var executeSQL *gorm.DB
	executeSQL = p.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

func (p promAlarmPageDo) Debug() IPromAlarmPageDo {
	return p.withDO(p.DO.Debug())
}

func (p promAlarmPageDo) WithContext(ctx context.Context) IPromAlarmPageDo {
	return p.withDO(p.DO.WithContext(ctx))
}

func (p promAlarmPageDo) ReadDB() IPromAlarmPageDo {
	return p.Clauses(dbresolver.Read)
}

func (p promAlarmPageDo) WriteDB() IPromAlarmPageDo {
	return p.Clauses(dbresolver.Write)
}

func (p promAlarmPageDo) Session(config *gorm.Session) IPromAlarmPageDo {
	return p.withDO(p.DO.Session(config))
}

func (p promAlarmPageDo) Clauses(conds ...clause.Expression) IPromAlarmPageDo {
	return p.withDO(p.DO.Clauses(conds...))
}

func (p promAlarmPageDo) Returning(value interface{}, columns ...string) IPromAlarmPageDo {
	return p.withDO(p.DO.Returning(value, columns...))
}

func (p promAlarmPageDo) Not(conds ...gen.Condition) IPromAlarmPageDo {
	return p.withDO(p.DO.Not(conds...))
}

func (p promAlarmPageDo) Or(conds ...gen.Condition) IPromAlarmPageDo {
	return p.withDO(p.DO.Or(conds...))
}

func (p promAlarmPageDo) Select(conds ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.Select(conds...))
}

func (p promAlarmPageDo) Where(conds ...gen.Condition) IPromAlarmPageDo {
	return p.withDO(p.DO.Where(conds...))
}

func (p promAlarmPageDo) Order(conds ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.Order(conds...))
}

func (p promAlarmPageDo) Distinct(cols ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.Distinct(cols...))
}

func (p promAlarmPageDo) Omit(cols ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.Omit(cols...))
}

func (p promAlarmPageDo) Join(table schema.Tabler, on ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.Join(table, on...))
}

func (p promAlarmPageDo) LeftJoin(table schema.Tabler, on ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.LeftJoin(table, on...))
}

func (p promAlarmPageDo) RightJoin(table schema.Tabler, on ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.RightJoin(table, on...))
}

func (p promAlarmPageDo) Group(cols ...field.Expr) IPromAlarmPageDo {
	return p.withDO(p.DO.Group(cols...))
}

func (p promAlarmPageDo) Having(conds ...gen.Condition) IPromAlarmPageDo {
	return p.withDO(p.DO.Having(conds...))
}

func (p promAlarmPageDo) Limit(limit int) IPromAlarmPageDo {
	return p.withDO(p.DO.Limit(limit))
}

func (p promAlarmPageDo) Offset(offset int) IPromAlarmPageDo {
	return p.withDO(p.DO.Offset(offset))
}

func (p promAlarmPageDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IPromAlarmPageDo {
	return p.withDO(p.DO.Scopes(funcs...))
}

func (p promAlarmPageDo) Unscoped() IPromAlarmPageDo {
	return p.withDO(p.DO.Unscoped())
}

func (p promAlarmPageDo) Create(values ...*model.PromAlarmPage) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Create(values)
}

func (p promAlarmPageDo) CreateInBatches(values []*model.PromAlarmPage, batchSize int) error {
	return p.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (p promAlarmPageDo) Save(values ...*model.PromAlarmPage) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Save(values)
}

func (p promAlarmPageDo) First() (*model.PromAlarmPage, error) {
	if result, err := p.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.PromAlarmPage), nil
	}
}

func (p promAlarmPageDo) Take() (*model.PromAlarmPage, error) {
	if result, err := p.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.PromAlarmPage), nil
	}
}

func (p promAlarmPageDo) Last() (*model.PromAlarmPage, error) {
	if result, err := p.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.PromAlarmPage), nil
	}
}

func (p promAlarmPageDo) Find() ([]*model.PromAlarmPage, error) {
	result, err := p.DO.Find()
	return result.([]*model.PromAlarmPage), err
}

func (p promAlarmPageDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.PromAlarmPage, err error) {
	buf := make([]*model.PromAlarmPage, 0, batchSize)
	err = p.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (p promAlarmPageDo) FindInBatches(result *[]*model.PromAlarmPage, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return p.DO.FindInBatches(result, batchSize, fc)
}

func (p promAlarmPageDo) Attrs(attrs ...field.AssignExpr) IPromAlarmPageDo {
	return p.withDO(p.DO.Attrs(attrs...))
}

func (p promAlarmPageDo) Assign(attrs ...field.AssignExpr) IPromAlarmPageDo {
	return p.withDO(p.DO.Assign(attrs...))
}

func (p promAlarmPageDo) Joins(fields ...field.RelationField) IPromAlarmPageDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Joins(_f))
	}
	return &p
}

func (p promAlarmPageDo) Preload(fields ...field.RelationField) IPromAlarmPageDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Preload(_f))
	}
	return &p
}

func (p promAlarmPageDo) FirstOrInit() (*model.PromAlarmPage, error) {
	if result, err := p.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.PromAlarmPage), nil
	}
}

func (p promAlarmPageDo) FirstOrCreate() (*model.PromAlarmPage, error) {
	if result, err := p.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.PromAlarmPage), nil
	}
}

func (p promAlarmPageDo) FindByPage(offset int, limit int) (result []*model.PromAlarmPage, count int64, err error) {
	result, err = p.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = p.Offset(-1).Limit(-1).Count()
	return
}

func (p promAlarmPageDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = p.Count()
	if err != nil {
		return
	}

	err = p.Offset(offset).Limit(limit).Scan(result)
	return
}

func (p promAlarmPageDo) Scan(result interface{}) (err error) {
	return p.DO.Scan(result)
}

func (p promAlarmPageDo) Delete(models ...*model.PromAlarmPage) (result gen.ResultInfo, err error) {
	return p.DO.Delete(models)
}

func (p *promAlarmPageDo) withDO(do gen.Dao) *promAlarmPageDo {
	p.DO = *do.(*gen.DO)
	return p
}
