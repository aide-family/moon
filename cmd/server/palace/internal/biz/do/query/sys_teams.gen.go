// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
)

func newSysTeam(db *gorm.DB, opts ...gen.DOOption) sysTeam {
	_sysTeam := sysTeam{}

	_sysTeam.sysTeamDo.UseDB(db, opts...)
	_sysTeam.sysTeamDo.UseModel(&model.SysTeam{})

	tableName := _sysTeam.sysTeamDo.TableName()
	_sysTeam.ALL = field.NewAsterisk(tableName)
	_sysTeam.ID = field.NewUint32(tableName, "id")
	_sysTeam.CreatedAt = field.NewField(tableName, "created_at")
	_sysTeam.UpdatedAt = field.NewField(tableName, "updated_at")
	_sysTeam.DeletedAt = field.NewInt64(tableName, "deleted_at")
	_sysTeam.Name = field.NewString(tableName, "name")
	_sysTeam.Status = field.NewField(tableName, "status")
	_sysTeam.Remark = field.NewString(tableName, "remark")
	_sysTeam.Avatar = field.NewString(tableName, "avatar")
	_sysTeam.OwnerID = field.NewUint32(tableName, "owner_id")
	_sysTeam.CreatorID = field.NewUint32(tableName, "creator_id")
	_sysTeam.Owner = sysTeamHasOneOwner{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Owner", "model.SysUser"),
	}

	_sysTeam.Creator = sysTeamHasOneCreator{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Creator", "model.SysUser"),
	}

	_sysTeam.Members = sysTeamHasManyMembers{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Members", "model.SysTeamMember"),
		Member: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Members.Member", "model.SysUser"),
		},
		Team: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Members.Team", "model.SysTeam"),
		},
	}

	_sysTeam.fillFieldMap()

	return _sysTeam
}

type sysTeam struct {
	sysTeamDo

	ALL       field.Asterisk
	ID        field.Uint32
	CreatedAt field.Field // 创建时间
	UpdatedAt field.Field // 更新时间
	DeletedAt field.Int64
	Name      field.String // 团队空间名
	Status    field.Field  // 状态
	Remark    field.String // 备注
	Avatar    field.String // 头像
	OwnerID   field.Uint32 // 负责人
	CreatorID field.Uint32 // 创建者
	Owner     sysTeamHasOneOwner

	Creator sysTeamHasOneCreator

	Members sysTeamHasManyMembers

	fieldMap map[string]field.Expr
}

func (s sysTeam) Table(newTableName string) *sysTeam {
	s.sysTeamDo.UseTable(newTableName)
	return s.updateTableName(newTableName)
}

func (s sysTeam) As(alias string) *sysTeam {
	s.sysTeamDo.DO = *(s.sysTeamDo.As(alias).(*gen.DO))
	return s.updateTableName(alias)
}

func (s *sysTeam) updateTableName(table string) *sysTeam {
	s.ALL = field.NewAsterisk(table)
	s.ID = field.NewUint32(table, "id")
	s.CreatedAt = field.NewField(table, "created_at")
	s.UpdatedAt = field.NewField(table, "updated_at")
	s.DeletedAt = field.NewInt64(table, "deleted_at")
	s.Name = field.NewString(table, "name")
	s.Status = field.NewField(table, "status")
	s.Remark = field.NewString(table, "remark")
	s.Avatar = field.NewString(table, "avatar")
	s.OwnerID = field.NewUint32(table, "owner_id")
	s.CreatorID = field.NewUint32(table, "creator_id")

	s.fillFieldMap()

	return s
}

func (s *sysTeam) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := s.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (s *sysTeam) fillFieldMap() {
	s.fieldMap = make(map[string]field.Expr, 13)
	s.fieldMap["id"] = s.ID
	s.fieldMap["created_at"] = s.CreatedAt
	s.fieldMap["updated_at"] = s.UpdatedAt
	s.fieldMap["deleted_at"] = s.DeletedAt
	s.fieldMap["name"] = s.Name
	s.fieldMap["status"] = s.Status
	s.fieldMap["remark"] = s.Remark
	s.fieldMap["avatar"] = s.Avatar
	s.fieldMap["owner_id"] = s.OwnerID
	s.fieldMap["creator_id"] = s.CreatorID

}

func (s sysTeam) clone(db *gorm.DB) sysTeam {
	s.sysTeamDo.ReplaceConnPool(db.Statement.ConnPool)
	return s
}

func (s sysTeam) replaceDB(db *gorm.DB) sysTeam {
	s.sysTeamDo.ReplaceDB(db)
	return s
}

type sysTeamHasOneOwner struct {
	db *gorm.DB

	field.RelationField
}

func (a sysTeamHasOneOwner) Where(conds ...field.Expr) *sysTeamHasOneOwner {
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

func (a sysTeamHasOneOwner) WithContext(ctx context.Context) *sysTeamHasOneOwner {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a sysTeamHasOneOwner) Session(session *gorm.Session) *sysTeamHasOneOwner {
	a.db = a.db.Session(session)
	return &a
}

func (a sysTeamHasOneOwner) Model(m *model.SysTeam) *sysTeamHasOneOwnerTx {
	return &sysTeamHasOneOwnerTx{a.db.Model(m).Association(a.Name())}
}

type sysTeamHasOneOwnerTx struct{ tx *gorm.Association }

func (a sysTeamHasOneOwnerTx) Find() (result *model.SysUser, err error) {
	return result, a.tx.Find(&result)
}

func (a sysTeamHasOneOwnerTx) Append(values ...*model.SysUser) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a sysTeamHasOneOwnerTx) Replace(values ...*model.SysUser) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a sysTeamHasOneOwnerTx) Delete(values ...*model.SysUser) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a sysTeamHasOneOwnerTx) Clear() error {
	return a.tx.Clear()
}

func (a sysTeamHasOneOwnerTx) Count() int64 {
	return a.tx.Count()
}

type sysTeamHasOneCreator struct {
	db *gorm.DB

	field.RelationField
}

func (a sysTeamHasOneCreator) Where(conds ...field.Expr) *sysTeamHasOneCreator {
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

func (a sysTeamHasOneCreator) WithContext(ctx context.Context) *sysTeamHasOneCreator {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a sysTeamHasOneCreator) Session(session *gorm.Session) *sysTeamHasOneCreator {
	a.db = a.db.Session(session)
	return &a
}

func (a sysTeamHasOneCreator) Model(m *model.SysTeam) *sysTeamHasOneCreatorTx {
	return &sysTeamHasOneCreatorTx{a.db.Model(m).Association(a.Name())}
}

type sysTeamHasOneCreatorTx struct{ tx *gorm.Association }

func (a sysTeamHasOneCreatorTx) Find() (result *model.SysUser, err error) {
	return result, a.tx.Find(&result)
}

func (a sysTeamHasOneCreatorTx) Append(values ...*model.SysUser) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a sysTeamHasOneCreatorTx) Replace(values ...*model.SysUser) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a sysTeamHasOneCreatorTx) Delete(values ...*model.SysUser) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a sysTeamHasOneCreatorTx) Clear() error {
	return a.tx.Clear()
}

func (a sysTeamHasOneCreatorTx) Count() int64 {
	return a.tx.Count()
}

type sysTeamHasManyMembers struct {
	db *gorm.DB

	field.RelationField

	Member struct {
		field.RelationField
	}
	Team struct {
		field.RelationField
	}
}

func (a sysTeamHasManyMembers) Where(conds ...field.Expr) *sysTeamHasManyMembers {
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

func (a sysTeamHasManyMembers) WithContext(ctx context.Context) *sysTeamHasManyMembers {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a sysTeamHasManyMembers) Session(session *gorm.Session) *sysTeamHasManyMembers {
	a.db = a.db.Session(session)
	return &a
}

func (a sysTeamHasManyMembers) Model(m *model.SysTeam) *sysTeamHasManyMembersTx {
	return &sysTeamHasManyMembersTx{a.db.Model(m).Association(a.Name())}
}

type sysTeamHasManyMembersTx struct{ tx *gorm.Association }

func (a sysTeamHasManyMembersTx) Find() (result []*model.SysTeamMember, err error) {
	return result, a.tx.Find(&result)
}

func (a sysTeamHasManyMembersTx) Append(values ...*model.SysTeamMember) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a sysTeamHasManyMembersTx) Replace(values ...*model.SysTeamMember) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a sysTeamHasManyMembersTx) Delete(values ...*model.SysTeamMember) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a sysTeamHasManyMembersTx) Clear() error {
	return a.tx.Clear()
}

func (a sysTeamHasManyMembersTx) Count() int64 {
	return a.tx.Count()
}

type sysTeamDo struct{ gen.DO }

type ISysTeamDo interface {
	gen.SubQuery
	Debug() ISysTeamDo
	WithContext(ctx context.Context) ISysTeamDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ISysTeamDo
	WriteDB() ISysTeamDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ISysTeamDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ISysTeamDo
	Not(conds ...gen.Condition) ISysTeamDo
	Or(conds ...gen.Condition) ISysTeamDo
	Select(conds ...field.Expr) ISysTeamDo
	Where(conds ...gen.Condition) ISysTeamDo
	Order(conds ...field.Expr) ISysTeamDo
	Distinct(cols ...field.Expr) ISysTeamDo
	Omit(cols ...field.Expr) ISysTeamDo
	Join(table schema.Tabler, on ...field.Expr) ISysTeamDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ISysTeamDo
	RightJoin(table schema.Tabler, on ...field.Expr) ISysTeamDo
	Group(cols ...field.Expr) ISysTeamDo
	Having(conds ...gen.Condition) ISysTeamDo
	Limit(limit int) ISysTeamDo
	Offset(offset int) ISysTeamDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ISysTeamDo
	Unscoped() ISysTeamDo
	Create(values ...*model.SysTeam) error
	CreateInBatches(values []*model.SysTeam, batchSize int) error
	Save(values ...*model.SysTeam) error
	First() (*model.SysTeam, error)
	Take() (*model.SysTeam, error)
	Last() (*model.SysTeam, error)
	Find() ([]*model.SysTeam, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.SysTeam, err error)
	FindInBatches(result *[]*model.SysTeam, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.SysTeam) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ISysTeamDo
	Assign(attrs ...field.AssignExpr) ISysTeamDo
	Joins(fields ...field.RelationField) ISysTeamDo
	Preload(fields ...field.RelationField) ISysTeamDo
	FirstOrInit() (*model.SysTeam, error)
	FirstOrCreate() (*model.SysTeam, error)
	FindByPage(offset int, limit int) (result []*model.SysTeam, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ISysTeamDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (s sysTeamDo) Debug() ISysTeamDo {
	return s.withDO(s.DO.Debug())
}

func (s sysTeamDo) WithContext(ctx context.Context) ISysTeamDo {
	return s.withDO(s.DO.WithContext(ctx))
}

func (s sysTeamDo) ReadDB() ISysTeamDo {
	return s.Clauses(dbresolver.Read)
}

func (s sysTeamDo) WriteDB() ISysTeamDo {
	return s.Clauses(dbresolver.Write)
}

func (s sysTeamDo) Session(config *gorm.Session) ISysTeamDo {
	return s.withDO(s.DO.Session(config))
}

func (s sysTeamDo) Clauses(conds ...clause.Expression) ISysTeamDo {
	return s.withDO(s.DO.Clauses(conds...))
}

func (s sysTeamDo) Returning(value interface{}, columns ...string) ISysTeamDo {
	return s.withDO(s.DO.Returning(value, columns...))
}

func (s sysTeamDo) Not(conds ...gen.Condition) ISysTeamDo {
	return s.withDO(s.DO.Not(conds...))
}

func (s sysTeamDo) Or(conds ...gen.Condition) ISysTeamDo {
	return s.withDO(s.DO.Or(conds...))
}

func (s sysTeamDo) Select(conds ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.Select(conds...))
}

func (s sysTeamDo) Where(conds ...gen.Condition) ISysTeamDo {
	return s.withDO(s.DO.Where(conds...))
}

func (s sysTeamDo) Order(conds ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.Order(conds...))
}

func (s sysTeamDo) Distinct(cols ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.Distinct(cols...))
}

func (s sysTeamDo) Omit(cols ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.Omit(cols...))
}

func (s sysTeamDo) Join(table schema.Tabler, on ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.Join(table, on...))
}

func (s sysTeamDo) LeftJoin(table schema.Tabler, on ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.LeftJoin(table, on...))
}

func (s sysTeamDo) RightJoin(table schema.Tabler, on ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.RightJoin(table, on...))
}

func (s sysTeamDo) Group(cols ...field.Expr) ISysTeamDo {
	return s.withDO(s.DO.Group(cols...))
}

func (s sysTeamDo) Having(conds ...gen.Condition) ISysTeamDo {
	return s.withDO(s.DO.Having(conds...))
}

func (s sysTeamDo) Limit(limit int) ISysTeamDo {
	return s.withDO(s.DO.Limit(limit))
}

func (s sysTeamDo) Offset(offset int) ISysTeamDo {
	return s.withDO(s.DO.Offset(offset))
}

func (s sysTeamDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ISysTeamDo {
	return s.withDO(s.DO.Scopes(funcs...))
}

func (s sysTeamDo) Unscoped() ISysTeamDo {
	return s.withDO(s.DO.Unscoped())
}

func (s sysTeamDo) Create(values ...*model.SysTeam) error {
	if len(values) == 0 {
		return nil
	}
	return s.DO.Create(values)
}

func (s sysTeamDo) CreateInBatches(values []*model.SysTeam, batchSize int) error {
	return s.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (s sysTeamDo) Save(values ...*model.SysTeam) error {
	if len(values) == 0 {
		return nil
	}
	return s.DO.Save(values)
}

func (s sysTeamDo) First() (*model.SysTeam, error) {
	if result, err := s.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.SysTeam), nil
	}
}

func (s sysTeamDo) Take() (*model.SysTeam, error) {
	if result, err := s.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.SysTeam), nil
	}
}

func (s sysTeamDo) Last() (*model.SysTeam, error) {
	if result, err := s.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.SysTeam), nil
	}
}

func (s sysTeamDo) Find() ([]*model.SysTeam, error) {
	result, err := s.DO.Find()
	return result.([]*model.SysTeam), err
}

func (s sysTeamDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.SysTeam, err error) {
	buf := make([]*model.SysTeam, 0, batchSize)
	err = s.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (s sysTeamDo) FindInBatches(result *[]*model.SysTeam, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return s.DO.FindInBatches(result, batchSize, fc)
}

func (s sysTeamDo) Attrs(attrs ...field.AssignExpr) ISysTeamDo {
	return s.withDO(s.DO.Attrs(attrs...))
}

func (s sysTeamDo) Assign(attrs ...field.AssignExpr) ISysTeamDo {
	return s.withDO(s.DO.Assign(attrs...))
}

func (s sysTeamDo) Joins(fields ...field.RelationField) ISysTeamDo {
	for _, _f := range fields {
		s = *s.withDO(s.DO.Joins(_f))
	}
	return &s
}

func (s sysTeamDo) Preload(fields ...field.RelationField) ISysTeamDo {
	for _, _f := range fields {
		s = *s.withDO(s.DO.Preload(_f))
	}
	return &s
}

func (s sysTeamDo) FirstOrInit() (*model.SysTeam, error) {
	if result, err := s.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.SysTeam), nil
	}
}

func (s sysTeamDo) FirstOrCreate() (*model.SysTeam, error) {
	if result, err := s.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.SysTeam), nil
	}
}

func (s sysTeamDo) FindByPage(offset int, limit int) (result []*model.SysTeam, count int64, err error) {
	result, err = s.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = s.Offset(-1).Limit(-1).Count()
	return
}

func (s sysTeamDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = s.Count()
	if err != nil {
		return
	}

	err = s.Offset(offset).Limit(limit).Scan(result)
	return
}

func (s sysTeamDo) Scan(result interface{}) (err error) {
	return s.DO.Scan(result)
}

func (s sysTeamDo) Delete(models ...*model.SysTeam) (result gen.ResultInfo, err error) {
	return s.DO.Delete(models)
}

func (s *sysTeamDo) withDO(do gen.Dao) *sysTeamDo {
	s.DO = *do.(*gen.DO)
	return s
}
