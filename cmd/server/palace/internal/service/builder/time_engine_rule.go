package builder

import (
	"context"
	"strconv"
	"strings"

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
		// Do 执行
		Do() ITimeEngineRuleDoBuilder
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

	timeEngineRuleDoBuilderImpl struct {
		ctx context.Context
	}

	timeEngineRuleModuleBuilderImpl struct {
		ctx context.Context
	}
)

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
		Rule: strings.Join(types.SliceTo(b.GetRule(), func(v int32) string {
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
		ID:       uint32(b.GetId()),
		Name:     b.GetData().GetName(),
		Remark:   b.GetData().GetRemark(),
		Status:   vobj.Status(b.GetData().GetStatus()),
		Category: vobj.TimeEngineRuleType(b.GetData().GetCategory()),
		Rule: strings.Join(types.SliceTo(b.GetData().GetRule(), func(v int32) string {
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
		ID: uint32(b.GetId()),
	}
}

// ToBo implements IGetTimeEngineRuleRequestBuilder.
func (b *getTimeEngineRuleRequestBuilderImpl) ToBo() *bo.GetTimeEngineRuleRequest {
	if b == nil || b.GetTimeEngineRuleRequest == nil {
		return nil
	}
	return &bo.GetTimeEngineRuleRequest{
		ID: uint32(b.GetId()),
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
		Id:       uint32(timeEngineRule.ID),
		Name:     timeEngineRule.Name,
		Category: int32(timeEngineRule.Category),
		Rule: types.SliceTo(strings.Split(timeEngineRule.Rule, ","), func(v string) int32 {
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
