package biz

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/msg"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/aide-family/moon/pkg/util/times"
)

type NotifyBiz struct {
	log *log.Helper

	notifyRepo   repository.NotifyRepo
	msgRepo      repository.MsgRepo
	strategyRepo repository.StrategyRepo
	userRepo     repository.UserRepo
	logX         repository.SysLogRepo
}

func NewNotifyBiz(
	repo repository.NotifyRepo,
	logX repository.SysLogRepo,
	msgRepo repository.MsgRepo,
	strategyRepo repository.StrategyRepo,
	userRepo repository.UserRepo,
	logger log.Logger,
) *NotifyBiz {
	return &NotifyBiz{
		log:          log.NewHelper(log.With(logger, "module", "biz.NotifyBiz")),
		notifyRepo:   repo,
		logX:         logX,
		msgRepo:      msgRepo,
		strategyRepo: strategyRepo,
		userRepo:     userRepo,
	}
}

// CreateNotify 创建通知对象
func (b *NotifyBiz) CreateNotify(ctx context.Context, notifyBo *bo.NotifyBO) (*bo.NotifyBO, error) {
	notifyBo, err := b.notifyRepo.Create(ctx, notifyBo)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyGroup,
		ModuleId:   notifyBo.Id,
		Content:    notifyBo.String(),
		Title:      "创建通知对象",
	})

	return notifyBo, nil
}

// CheckNotifyName 检查通知名称是否存在
func (b *NotifyBiz) CheckNotifyName(ctx context.Context, name string, id ...uint32) error {
	total, err := b.notifyRepo.Count(ctx, basescopes.NameEQ(name), basescopes.NotInIds(id...))
	if err != nil {
		return err
	}
	if total > 0 {
		return perrors.ErrorAlreadyExists("通知对象名称已存在")
	}

	return nil
}

// UpdateNotifyById 更新通知对象
func (b *NotifyBiz) UpdateNotifyById(ctx context.Context, id uint32, notifyBo *bo.NotifyBO) error {
	// 查询
	oldData, err := b.GetNotifyById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return perrors.ErrorNotFound("通知对象不存在")
		}
		return err
	}
	if err = b.notifyRepo.Update(ctx, notifyBo, basescopes.InIds(id)); err != nil {
		return err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyGroup,
		ModuleId:   id,
		Content:    bo.NewChangeLogBo(oldData, notifyBo).String(),
		Title:      "更新通知对象",
	})
	return nil
}

// DeleteNotifyById 删除通知对象
func (b *NotifyBiz) DeleteNotifyById(ctx context.Context, id uint32) error {
	// 查询
	oldData, err := b.GetNotifyById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return perrors.ErrorNotFound("通知对象不存在")
		}
		return err
	}
	if err = b.notifyRepo.Delete(ctx, basescopes.InIds(id)); err != nil {
		return err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionDelete, &bo.SysLogBo{
		ModuleName: vobj.ModuleAlarmNotifyGroup,
		ModuleId:   id,
		Content:    oldData.String(),
		Title:      "删除通知对象",
	})
	return nil
}

// GetNotifyById 获取通知对象
func (b *NotifyBiz) GetNotifyById(ctx context.Context, id uint32) (*bo.NotifyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.InIds(id),
		do.PromAlarmNotifyPreloadChatGroups(),
		do.PromAlarmNotifyPreloadBeNotifyMembers(),
	}
	notifyBo, err := b.notifyRepo.Get(ctx, wheres...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perrors.ErrorNotFound("通知对象不存在")
		}
		return nil, err
	}

	return notifyBo, nil
}

// ListNotify 获取通知对象列表
func (b *NotifyBiz) ListNotify(ctx context.Context, req *bo.ListNotifyRequest) ([]*bo.NotifyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(req.Keyword),
		basescopes.StatusEQ(req.Status),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		do.PromAlarmNotifyPreloadChatGroups(),
		do.PromAlarmNotifyPreloadBeNotifyMembers(),
	}
	notifyBos, err := b.notifyRepo.List(ctx, req.Page, wheres...)
	if err != nil {
		return nil, err
	}

	return notifyBos, nil
}

// TestNotifyTemplate 测试告警消息
func (b *NotifyBiz) TestNotifyTemplate(ctx context.Context, req *bo.TestNotifyTemplateParams) error {
	// 查询策略详情， 获取通知对象
	strategyDetail, err := b.strategyRepo.GetStrategyById(ctx, req.StrategyId, do.StrategyPreloadPromNotifies(do.PromAlarmNotifyPreloadFieldChatGroups))
	if err != nil {
		return err
	}
	notifies := strategyDetail.GetPromNotifies()
	if strategyDetail == nil || len(notifies) == 0 {
		return perrors.ErrorNotFound("未配置告警对象，请先配置后再测试告警")
	}
	hooks := make([]*bo.ChatGroupBO, 0, len(notifies))
	for _, notify := range notifies {
		if len(notify.GetChatGroups()) == 0 {
			continue
		}
		hooks = append(hooks, notify.GetChatGroups()...)
	}

	now := time.Now()
	hookMsg := &msg.HookNotifyMsg{
		Content: req.Template,
		AlarmInfo: &bo.AlertBo{
			Status: vobj.AlarmStatusAlarm.EN(),
			Labels: &strategy.Labels{
				strategy.MetricInstance: "localhost",
				strategy.MetricAlert:    "test_alert",
				"endpoint":              "127.0.0.1",
				"job":                   "test",
				"severity":              "critical",
				"app":                   "moon",
			},
			Annotations: &strategy.Annotations{
				strategy.MetricSummary:     "test hook template summary",
				strategy.MetricDescription: "test hook template description",
			},
			StartsAt:     now.Add(-time.Minute * 5).Format(times.ParseLayout),
			EndsAt:       now.Format(times.ParseLayout),
			GeneratorURL: "https://github.com/aide-family/moon",
			Fingerprint:  hash.MD5(now.String()),
		},
	}
	dataMap := hookMsg.AlarmInfo.ToMap()
	dataMap["value"] = 100
	hookMsg.Content = strategy.Formatter(hookMsg.Content, dataMap)

	eg := new(errgroup.Group)
	switch req.NotifyType {
	case vobj.NotifyTemplateTypeCustom:
		customHooks := getHookItemList(hooks, vobj.NotifyAppCustom)
		eg.Go(func() error {
			return notifyHooks(ctx, customHooks, vobj.NotifyAppCustom, hookMsg)
		})
	case vobj.NotifyTemplateTypeWeChatWork:
		wechatWorkHooks := getHookItemList(hooks, vobj.NotifyAppWeChatWork)
		eg.Go(func() error {
			return notifyHooks(ctx, wechatWorkHooks, vobj.NotifyAppWeChatWork, hookMsg)
		})
	case vobj.NotifyTemplateTypeDingDing:
		dingdingHooks := getHookItemList(hooks, vobj.NotifyAppDingDing)
		eg.Go(func() error {
			return notifyHooks(ctx, dingdingHooks, vobj.NotifyAppDingDing, hookMsg)
		})
	case vobj.NotifyTemplateTypeFeiShu:
		feiShuHooks := getHookItemList(hooks, vobj.NotifyAppFeiShu)
		eg.Go(func() error {
			return notifyHooks(ctx, feiShuHooks, vobj.NotifyAppFeiShu, hookMsg)
		})
	case vobj.NotifyTemplateTypeEmail, vobj.NotifyTemplateTypeSms:
		// 查询用户信息
		userId := middler.GetUserId(ctx)
		user, err := b.userRepo.Get(ctx, basescopes.InIds(userId))
		if err != nil {
			return perrors.ErrorNotFound("用户不存在")
		}
		memberNotifyType := vobj.NotifyTypeEmail
		if vobj.NotifyTemplateTypeSms == req.NotifyType {
			memberNotifyType = vobj.NotifyTypeSms
		}
		b.log.Debugw("user", user)
		members := []*bo.NotifyMemberBO{
			{
				MemberId:   userId,
				Member:     user,
				NotifyType: memberNotifyType,
			},
		}
		memberTemplateMap := map[vobj.NotifyType]string{
			memberNotifyType: req.Template,
		}
		eg.Go(func() error {
			return b.msgRepo.SendAlarmToMember(ctx, members, memberTemplateMap, hookMsg.AlarmInfo)
		})
	default:
		return perrors.ErrorNotFound("未知的通知方式")
	}
	return eg.Wait()
}

func getHookItemList(hooks []*bo.ChatGroupBO, notifyType vobj.NotifyApp) []*bo.ChatGroupBO {
	customHooks := make([]*bo.ChatGroupBO, 0, len(hooks))
	for _, hook := range hooks {
		if hook.NotifyApp == notifyType {
			customHooks = append(customHooks, hook)
		}
	}
	return customHooks
}

func notifyHooks(ctx context.Context, hooks []*bo.ChatGroupBO, notifyType vobj.NotifyApp, message *msg.HookNotifyMsg) error {
	if len(hooks) == 0 {
		return perrors.ErrorNotFound("你没有对应的hook对象，请先绑定后再重试")
	}
	eg := new(errgroup.Group)
	for _, hook := range hooks {
		hookTemp := hook
		messageTmp := message
		messageTmp.Secret = hookTemp.Secret
		eg.Go(func() error {
			return msg.NewHookNotify(notifyType).Alarm(ctx, hookTemp.Hook, message)
		})
	}
	return eg.Wait()
}
