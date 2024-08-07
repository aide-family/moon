package build

import (
	"context"

	adminapi "github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

func newAlarmPageModuleBuilder(ctx context.Context) AlarmPageModuleBuilder {
	return &alarmPageModelBuilder{
		ctx: ctx,
	}
}

func newAlarmPageBuilder(ctx context.Context, alarmPage *bizmodel.AlarmPageSelf) AlarmPageBuilder {
	return &alarmPageBuilder{
		ctx:       ctx,
		alarmPage: alarmPage,
	}
}

func newAlarmPagesBuilder(ctx context.Context, alarmPages []*bizmodel.AlarmPageSelf) AlarmPagesBuilder {
	return &alarmPagesBuilder{
		ctx:        ctx,
		alarmPages: alarmPages,
	}
}

type (
	// AlarmPageModuleBuilder 告警页面构造器
	AlarmPageModuleBuilder interface {
		WithAlarmPages([]*bizmodel.AlarmPageSelf) AlarmPagesBuilder
		WithAlarmPage(*bizmodel.AlarmPageSelf) AlarmPageBuilder
	}

	// AlarmPageBuilder 告警页面构造器
	AlarmPageBuilder interface {
		ToAPI() *adminapi.SelfAlarmPageItem
	}

	// AlarmPagesBuilder 告警页面构造器
	AlarmPagesBuilder interface {
		ToAPIs() []*adminapi.SelfAlarmPageItem
	}

	alarmPageModelBuilder struct {
		ctx context.Context
	}

	alarmPageBuilder struct {
		ctx       context.Context
		alarmPage *bizmodel.AlarmPageSelf
	}

	alarmPagesBuilder struct {
		ctx        context.Context
		alarmPages []*bizmodel.AlarmPageSelf
	}
)

func (a *alarmPageModelBuilder) WithAlarmPages(selves []*bizmodel.AlarmPageSelf) AlarmPagesBuilder {
	return newAlarmPagesBuilder(a.ctx, selves)
}

func (a *alarmPageModelBuilder) WithAlarmPage(self *bizmodel.AlarmPageSelf) AlarmPageBuilder {
	return newAlarmPageBuilder(a.ctx, self)
}

func (a *alarmPagesBuilder) ToAPIs() []*adminapi.SelfAlarmPageItem {
	if types.IsNil(a) || len(a.alarmPages) == 0 {
		return nil
	}

	return types.SliceToWithFilter(a.alarmPages, func(alarmPage *bizmodel.AlarmPageSelf) (*adminapi.SelfAlarmPageItem, bool) {
		item := newAlarmPageBuilder(a.ctx, alarmPage).ToAPI()
		if types.IsNil(item) {
			return nil, false
		}
		return item, true
	})
}

func (a *alarmPageBuilder) ToAPI() *adminapi.SelfAlarmPageItem {
	if types.IsNil(a) || types.IsNil(a.alarmPage) || types.IsNil(a.alarmPage.AlarmPage) {
		return nil
	}
	alarmPage := a.alarmPage.AlarmPage
	return &adminapi.SelfAlarmPageItem{
		Id:           alarmPage.GetID(),
		Name:         alarmPage.GetName(),
		ColorType:    alarmPage.GetColorType(),
		CssClass:     alarmPage.GetCSSClass(),
		Value:        alarmPage.GetValue(),
		Icon:         alarmPage.GetIcon(),
		ImageUrl:     alarmPage.GetImageURL(),
		LanguageCode: alarmPage.GetLanguageCode(),
		Remark:       alarmPage.GetRemark(),
	}
}
