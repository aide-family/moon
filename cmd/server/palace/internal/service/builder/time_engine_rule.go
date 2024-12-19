package builder

import (
	"context"
	"strconv"
	"strings"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	alarmapi "github.com/aide-family/moon/api/admin/alarm"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// ITimeEngineRuleModuleBuilder 时间引擎规则模块构造器
	ITimeEngineRuleModuleBuilder interface {
		// WithCreateTimeEngineRuleRequest 创建时间引擎规则请求构造器
		WithCreateTimeEngineRuleRequest(req *alarmapi.CreateTimeEngineRuleRequest) ICreateTimeEngineRuleRequestBuilder
		// WithUpdateTimeEngineRuleRequest 更新时间引擎规则请求构造器
		WithUpdateTimeEngineRuleRequest(req *alarmapi.UpdateTimeEngineRuleRequest) IUpdateTimeEngineRuleRequestBuilder
		// WithDeleteTimeEngineRuleRequest 删除时间引擎规则请求构造器
		WithDeleteTimeEngineRuleRequest(req *alarmapi.DeleteTimeEngineRuleRequest) IDeleteTimeEngineRuleRequestBuilder
		// WithGetTimeEngineRuleRequest 获取时间引擎规则请求构造器
		WithGetTimeEngineRuleRequest(req *alarmapi.GetTimeEngineRuleRequest) IGetTimeEngineRuleRequestBuilder
		// WithListTimeEngineRuleRequest 获取时间引擎规则列表请求构造器
		WithListTimeEngineRuleRequest(req *alarmapi.ListTimeEngineRuleRequest) IListTimeEngineRuleRequestBuilder
		// WithBatchUpdateTimeEngineRuleStatusRequest 批量更新时间引擎规则状态请求构造器
		WithBatchUpdateTimeEngineRuleStatusRequest(req *alarmapi.BatchUpdateTimeEngineRuleStatusRequest) IBatchUpdateTimeEngineRuleStatusRequestBuilder
		// Do 执行
		Do() ITimeEngineRuleDoBuilder
	}

	// ITimeEngineModuleBuilder 时间引擎模块构造器
	ITimeEngineModuleBuilder interface {
		// WithCreateTimeEngineRequest 创建时间引擎请求构造器
		WithCreateTimeEngineRequest(req *alarmapi.CreateTimeEngineRequest) ICreateTimeEngineRequestBuilder
		// WithUpdateTimeEngineRequest 更新时间引擎请求构造器
		WithUpdateTimeEngineRequest(req *alarmapi.UpdateTimeEngineRequest) IUpdateTimeEngineRequestBuilder
		// WithDeleteTimeEngineRequest 删除时间引擎请求构造器
		WithDeleteTimeEngineRequest(req *alarmapi.DeleteTimeEngineRequest) IDeleteTimeEngineRequestBuilder
		// WithGetTimeEngineRequest 获取时间引擎请求构造器
		WithGetTimeEngineRequest(req *alarmapi.GetTimeEngineRequest) IGetTimeEngineRequestBuilder
		// WithListTimeEngineRequest 获取时间引擎列表请求构造器
		WithListTimeEngineRequest(req *alarmapi.ListTimeEngineRequest) IListTimeEngineRequestBuilder
		// WithBatchUpdateTimeEngineStatusRequest 批量更新时间引擎状态请求构造器
		WithBatchUpdateTimeEngineStatusRequest(req *alarmapi.BatchUpdateTimeEngineStatusRequest) IBatchUpdateTimeEngineStatusRequestBuilder
		// Do 执行
		Do() ITimeEngineDoBuilder
	}

	// ICreateTimeEngineRequestBuilder 创建时间引擎请求构造器
	ICreateTimeEngineRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.CreateTimeEngineRequest
	}

	// IUpdateTimeEngineRequestBuilder 更新时间引擎请求构造器
	IUpdateTimeEngineRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.UpdateTimeEngineRequest
	}

	// IDeleteTimeEngineRequestBuilder 删除时间引擎请求构造器
	IDeleteTimeEngineRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.DeleteTimeEngineRequest
	}

	// IGetTimeEngineRequestBuilder 获取时间引擎请求构造器
	IGetTimeEngineRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.GetTimeEngineRequest
	}

	// IListTimeEngineRequestBuilder 获取时间引擎列表请求构造器
	IListTimeEngineRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.ListTimeEngineRequest
	}

	// IBatchUpdateTimeEngineStatusRequestBuilder 批量更新时间引擎状态请求构造器
	IBatchUpdateTimeEngineStatusRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.BatchUpdateTimeEngineStatusRequest
	}

	// ITimeEngineDoBuilder 时间引擎执行构造器
	ITimeEngineDoBuilder interface {
		// ToAPI 转换为API
		ToAPI(*bizmodel.TimeEngine) *adminapi.TimeEngineItem
		// ToAPIs 转换为API列表
		ToAPIs([]*bizmodel.TimeEngine) []*adminapi.TimeEngineItem
	}

	// ICreateTimeEngineRuleRequestBuilder 创建时间引擎规则请求构造器
	ICreateTimeEngineRuleRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.CreateTimeEngineRuleRequest
	}

	// IUpdateTimeEngineRuleRequestBuilder 更新时间引擎规则请求构造器
	IUpdateTimeEngineRuleRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.UpdateTimeEngineRuleRequest
	}

	// IDeleteTimeEngineRuleRequestBuilder 删除时间引擎规则请求构造器
	IDeleteTimeEngineRuleRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.DeleteTimeEngineRuleRequest
	}

	// IGetTimeEngineRuleRequestBuilder 获取时间引擎规则请求构造器
	IGetTimeEngineRuleRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.GetTimeEngineRuleRequest
	}

	// IListTimeEngineRuleRequestBuilder 获取时间引擎规则列表请求构造器
	IListTimeEngineRuleRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.ListTimeEngineRuleRequest
	}

	// IBatchUpdateTimeEngineRuleStatusRequestBuilder 批量更新时间引擎规则状态请求构造器
	IBatchUpdateTimeEngineRuleStatusRequestBuilder interface {
		// ToBo 转换为BO
		ToBo() *bo.BatchUpdateTimeEngineRuleStatusRequest
	}

	// ITimeEngineRuleDoBuilder 时间引擎规则执行构造器
	ITimeEngineRuleDoBuilder interface {
		// ToAPI 转换为API
		ToAPI(*bizmodel.TimeEngineRule) *adminapi.TimeEngineRuleItem
		// ToAPIs 转换为API列表
		ToAPIs([]*bizmodel.TimeEngineRule) []*adminapi.TimeEngineRuleItem
	}

	createTimeEngineRuleRequestBuilderImpl struct {
		*alarmapi.CreateTimeEngineRuleRequest
		ctx context.Context
	}

	updateTimeEngineRuleRequestBuilderImpl struct {
		*alarmapi.UpdateTimeEngineRuleRequest
		ctx context.Context
	}

	deleteTimeEngineRuleRequestBuilderImpl struct {
		*alarmapi.DeleteTimeEngineRuleRequest
		ctx context.Context
	}

	getTimeEngineRuleRequestBuilderImpl struct {
		*alarmapi.GetTimeEngineRuleRequest
		ctx context.Context
	}

	listTimeEngineRuleRequestBuilderImpl struct {
		*alarmapi.ListTimeEngineRuleRequest
		ctx context.Context
	}

	batchUpdateTimeEngineRuleStatusRequestBuilderImpl struct {
		*alarmapi.BatchUpdateTimeEngineRuleStatusRequest
		ctx context.Context
	}

	timeEngineRuleDoBuilderImpl struct {
		ctx context.Context
	}

	timeEngineRuleModuleBuilderImpl struct {
		ctx context.Context
	}

	createTimeEngineRequestBuilderImpl struct {
		*alarmapi.CreateTimeEngineRequest
		ctx context.Context
	}

	updateTimeEngineRequestBuilderImpl struct {
		*alarmapi.UpdateTimeEngineRequest
		ctx context.Context
	}

	deleteTimeEngineRequestBuilderImpl struct {
		*alarmapi.DeleteTimeEngineRequest
		ctx context.Context
	}

	getTimeEngineRequestBuilderImpl struct {
		*alarmapi.GetTimeEngineRequest
		ctx context.Context
	}

	listTimeEngineRequestBuilderImpl struct {
		*alarmapi.ListTimeEngineRequest
		ctx context.Context
	}

	batchUpdateTimeEngineStatusRequestBuilderImpl struct {
		*alarmapi.BatchUpdateTimeEngineStatusRequest
		ctx context.Context
	}

	timeEngineDoBuilderImpl struct {
		ctx context.Context
	}

	timeEngineModuleBuilderImpl struct {
		ctx context.Context
	}
)

// ToBo implements ICreateTimeEngineRequestBuilder.
func (c *createTimeEngineRequestBuilderImpl) ToBo() *bo.CreateTimeEngineRequest {
	if c == nil || c.CreateTimeEngineRequest == nil {
		return nil
	}
	return &bo.CreateTimeEngineRequest{
		Name:    c.GetName(),
		Remark:  c.GetRemark(),
		Status:  vobj.Status(c.GetStatus()),
		RuleIDs: c.GetRules(),
	}
}

// ToBo implements IUpdateTimeEngineRequestBuilder.
func (u *updateTimeEngineRequestBuilderImpl) ToBo() *bo.UpdateTimeEngineRequest {
	if u == nil || u.UpdateTimeEngineRequest == nil {
		return nil
	}
	return &bo.UpdateTimeEngineRequest{
		ID:      u.GetId(),
		Name:    u.GetData().GetName(),
		Remark:  u.GetData().GetRemark(),
		Status:  vobj.Status(u.GetData().GetStatus()),
		RuleIDs: u.GetData().GetRules(),
	}
}

// ToBo implements IDeleteTimeEngineRequestBuilder.
func (d *deleteTimeEngineRequestBuilderImpl) ToBo() *bo.DeleteTimeEngineRequest {
	if d == nil || d.DeleteTimeEngineRequest == nil {
		return nil
	}
	return &bo.DeleteTimeEngineRequest{
		ID: d.GetId(),
	}
}

// ToBo implements IGetTimeEngineRequestBuilder.
func (g *getTimeEngineRequestBuilderImpl) ToBo() *bo.GetTimeEngineRequest {
	if g == nil || g.GetTimeEngineRequest == nil {
		return nil
	}
	return &bo.GetTimeEngineRequest{
		ID: g.GetId(),
	}
}

// ToBo implements IListTimeEngineRequestBuilder.
func (l *listTimeEngineRequestBuilderImpl) ToBo() *bo.ListTimeEngineRequest {
	if l == nil || l.ListTimeEngineRequest == nil {
		return nil
	}
	return &bo.ListTimeEngineRequest{
		Page:    types.NewPagination(l.GetPagination()),
		Status:  vobj.Status(l.GetStatus()),
		Keyword: l.GetKeyword(),
	}
}

// ToBo implements IBatchUpdateTimeEngineStatusRequestBuilder.
func (b *batchUpdateTimeEngineStatusRequestBuilderImpl) ToBo() *bo.BatchUpdateTimeEngineStatusRequest {
	if b == nil || b.BatchUpdateTimeEngineStatusRequest == nil {
		return nil
	}
	return &bo.BatchUpdateTimeEngineStatusRequest{
		IDs:    b.GetIds(),
		Status: vobj.Status(b.GetStatus()),
	}
}

// ToAPI implements ITimeEngineDoBuilder.
func (t *timeEngineDoBuilderImpl) ToAPI(timeEngine *bizmodel.TimeEngine) *adminapi.TimeEngineItem {
	if t == nil || timeEngine == nil {
		return nil
	}
	userMap := getUsers(t.ctx, nil, timeEngine.GetCreatorID())
	return &adminapi.TimeEngineItem{
		Id:        timeEngine.ID,
		Name:      timeEngine.Name,
		Status:    api.Status(timeEngine.Status),
		Remark:    timeEngine.Remark,
		CreatedAt: timeEngine.CreatedAt.Unix(),
		UpdatedAt: timeEngine.UpdatedAt.Unix(),
		Rules:     NewParamsBuild(t.ctx).TimeEngineRuleModuleBuilder().Do().ToAPIs(timeEngine.Rules),
		Creator:   userMap[timeEngine.GetCreatorID()],
	}
}

// ToAPIs implements ITimeEngineDoBuilder.
func (t *timeEngineDoBuilderImpl) ToAPIs(timeEngines []*bizmodel.TimeEngine) []*adminapi.TimeEngineItem {
	if t == nil || timeEngines == nil {
		return nil
	}
	return types.SliceTo(timeEngines, t.ToAPI)
}

// WithCreateTimeEngineRequest implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) WithCreateTimeEngineRequest(req *alarmapi.CreateTimeEngineRequest) ICreateTimeEngineRequestBuilder {
	return &createTimeEngineRequestBuilderImpl{CreateTimeEngineRequest: req, ctx: b.ctx}
}

// WithUpdateTimeEngineRequest implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) WithUpdateTimeEngineRequest(req *alarmapi.UpdateTimeEngineRequest) IUpdateTimeEngineRequestBuilder {
	return &updateTimeEngineRequestBuilderImpl{UpdateTimeEngineRequest: req, ctx: b.ctx}
}

// WithDeleteTimeEngineRequest implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) WithDeleteTimeEngineRequest(req *alarmapi.DeleteTimeEngineRequest) IDeleteTimeEngineRequestBuilder {
	return &deleteTimeEngineRequestBuilderImpl{DeleteTimeEngineRequest: req, ctx: b.ctx}
}

// WithGetTimeEngineRequest implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) WithGetTimeEngineRequest(req *alarmapi.GetTimeEngineRequest) IGetTimeEngineRequestBuilder {
	return &getTimeEngineRequestBuilderImpl{GetTimeEngineRequest: req, ctx: b.ctx}
}

// WithListTimeEngineRequest implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) WithListTimeEngineRequest(req *alarmapi.ListTimeEngineRequest) IListTimeEngineRequestBuilder {
	return &listTimeEngineRequestBuilderImpl{ListTimeEngineRequest: req, ctx: b.ctx}
}

// WithBatchUpdateTimeEngineStatusRequest implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) WithBatchUpdateTimeEngineStatusRequest(req *alarmapi.BatchUpdateTimeEngineStatusRequest) IBatchUpdateTimeEngineStatusRequestBuilder {
	return &batchUpdateTimeEngineStatusRequestBuilderImpl{BatchUpdateTimeEngineStatusRequest: req, ctx: b.ctx}
}

// Do implements ITimeEngineModuleBuilder.
func (b *timeEngineModuleBuilderImpl) Do() ITimeEngineDoBuilder {
	return &timeEngineDoBuilderImpl{ctx: b.ctx}
}

// WithCreateTimeEngineRuleRequest implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) WithCreateTimeEngineRuleRequest(req *alarmapi.CreateTimeEngineRuleRequest) ICreateTimeEngineRuleRequestBuilder {
	return &createTimeEngineRuleRequestBuilderImpl{CreateTimeEngineRuleRequest: req, ctx: b.ctx}
}

// WithUpdateTimeEngineRuleRequest implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) WithUpdateTimeEngineRuleRequest(req *alarmapi.UpdateTimeEngineRuleRequest) IUpdateTimeEngineRuleRequestBuilder {
	return &updateTimeEngineRuleRequestBuilderImpl{UpdateTimeEngineRuleRequest: req, ctx: b.ctx}
}

// WithDeleteTimeEngineRuleRequest implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) WithDeleteTimeEngineRuleRequest(req *alarmapi.DeleteTimeEngineRuleRequest) IDeleteTimeEngineRuleRequestBuilder {
	return &deleteTimeEngineRuleRequestBuilderImpl{DeleteTimeEngineRuleRequest: req, ctx: b.ctx}
}

// WithGetTimeEngineRuleRequest implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) WithGetTimeEngineRuleRequest(req *alarmapi.GetTimeEngineRuleRequest) IGetTimeEngineRuleRequestBuilder {
	return &getTimeEngineRuleRequestBuilderImpl{GetTimeEngineRuleRequest: req, ctx: b.ctx}
}

// WithListTimeEngineRuleRequest implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) WithListTimeEngineRuleRequest(req *alarmapi.ListTimeEngineRuleRequest) IListTimeEngineRuleRequestBuilder {
	return &listTimeEngineRuleRequestBuilderImpl{ListTimeEngineRuleRequest: req, ctx: b.ctx}
}

// WithBatchUpdateTimeEngineRuleStatusRequest implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) WithBatchUpdateTimeEngineRuleStatusRequest(req *alarmapi.BatchUpdateTimeEngineRuleStatusRequest) IBatchUpdateTimeEngineRuleStatusRequestBuilder {
	return &batchUpdateTimeEngineRuleStatusRequestBuilderImpl{BatchUpdateTimeEngineRuleStatusRequest: req, ctx: b.ctx}
}

// Do implements ITimeEngineRuleModuleBuilder.
func (b *timeEngineRuleModuleBuilderImpl) Do() ITimeEngineRuleDoBuilder {
	return &timeEngineRuleDoBuilderImpl{ctx: b.ctx}
}

// ToBo implements ICreateTimeEngineRuleRequestBuilder.
func (b *createTimeEngineRuleRequestBuilderImpl) ToBo() *bo.CreateTimeEngineRuleRequest {
	if b == nil || b.CreateTimeEngineRuleRequest == nil {
		return nil
	}
	return &bo.CreateTimeEngineRuleRequest{
		Name:     b.GetName(),
		Remark:   b.GetRemark(),
		Status:   vobj.Status(b.GetStatus()),
		Category: vobj.TimeEngineRuleType(b.GetCategory()),
		Rule: strings.Join(types.SliceTo(b.GetRules(), func(v int32) string {
			return strconv.Itoa(int(v))
		}), ","),
	}
}

// ToBo implements IUpdateTimeEngineRuleRequestBuilder.
func (b *updateTimeEngineRuleRequestBuilderImpl) ToBo() *bo.UpdateTimeEngineRuleRequest {
	if b == nil || b.UpdateTimeEngineRuleRequest == nil {
		return nil
	}
	return &bo.UpdateTimeEngineRuleRequest{
		ID:       b.GetId(),
		Name:     b.GetData().GetName(),
		Remark:   b.GetData().GetRemark(),
		Status:   vobj.Status(b.GetData().GetStatus()),
		Category: vobj.TimeEngineRuleType(b.GetData().GetCategory()),
		Rule: strings.Join(types.SliceTo(b.GetData().GetRules(), func(v int32) string {
			return strconv.Itoa(int(v))
		}), ","),
	}
}

// ToBo implements IDeleteTimeEngineRuleRequestBuilder.
func (b *deleteTimeEngineRuleRequestBuilderImpl) ToBo() *bo.DeleteTimeEngineRuleRequest {
	if b == nil || b.DeleteTimeEngineRuleRequest == nil {
		return nil
	}
	return &bo.DeleteTimeEngineRuleRequest{
		ID: b.GetId(),
	}
}

// ToBo implements IGetTimeEngineRuleRequestBuilder.
func (b *getTimeEngineRuleRequestBuilderImpl) ToBo() *bo.GetTimeEngineRuleRequest {
	if b == nil || b.GetTimeEngineRuleRequest == nil {
		return nil
	}
	return &bo.GetTimeEngineRuleRequest{
		ID: b.GetId(),
	}
}

// ToBo implements IListTimeEngineRuleRequestBuilder.
func (b *listTimeEngineRuleRequestBuilderImpl) ToBo() *bo.ListTimeEngineRuleRequest {
	if b == nil || b.ListTimeEngineRuleRequest == nil {
		return nil
	}
	return &bo.ListTimeEngineRuleRequest{
		Page:     types.NewPagination(b.GetPagination()),
		Category: vobj.TimeEngineRuleType(b.GetCategory()),
		Status:   vobj.Status(b.GetStatus()),
		Keyword:  b.GetKeyword(),
	}
}

// ToAPI implements ITimeEngineRuleDoBuilder.
func (b *timeEngineRuleDoBuilderImpl) ToAPI(timeEngineRule *bizmodel.TimeEngineRule) *adminapi.TimeEngineRuleItem {
	if b == nil || timeEngineRule == nil {
		return nil
	}
	userMap := getUsers(b.ctx, nil, timeEngineRule.GetCreatorID())
	return &adminapi.TimeEngineRuleItem{
		Id:       timeEngineRule.ID,
		Name:     timeEngineRule.Name,
		Category: int32(timeEngineRule.Category),
		Rules: types.SliceTo(strings.Split(timeEngineRule.Rule, ","), func(v string) int32 {
			n, err := strconv.Atoi(v)
			if err != nil {
				return 0
			}
			return int32(n)
		}),
		Status:    int32(timeEngineRule.Status),
		Remark:    timeEngineRule.Remark,
		CreatedAt: timeEngineRule.CreatedAt.Unix(),
		UpdatedAt: timeEngineRule.UpdatedAt.Unix(),
		Creator:   userMap[timeEngineRule.GetCreatorID()],
	}
}

// ToAPIs implements ITimeEngineRuleDoBuilder.
func (b *timeEngineRuleDoBuilderImpl) ToAPIs(timeEngineRules []*bizmodel.TimeEngineRule) []*adminapi.TimeEngineRuleItem {
	if b == nil || timeEngineRules == nil {
		return nil
	}
	return types.SliceTo(timeEngineRules, b.ToAPI)
}

// ToBo implements IBatchUpdateTimeEngineRuleStatusRequestBuilder.
func (b *batchUpdateTimeEngineRuleStatusRequestBuilderImpl) ToBo() *bo.BatchUpdateTimeEngineRuleStatusRequest {
	if b == nil || b.BatchUpdateTimeEngineRuleStatusRequest == nil {
		return nil
	}
	return &bo.BatchUpdateTimeEngineRuleStatusRequest{
		IDs:    b.GetIds(),
		Status: vobj.Status(b.GetStatus()),
	}
}
