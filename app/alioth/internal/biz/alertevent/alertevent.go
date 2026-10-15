package alertevent

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// EvalInterval 是调度周期，与调度锁的 TTL 相同。
	EvalInterval = 15 * time.Second
	// ScheduleLockKey 是到期扫描使用的锁。玉衡不持有评测锁。
	ScheduleLockKey = "alioth:lock:schedule:eval"
)

var ProviderSet = wire.NewSet(NewUsecase)

// Usecase 调度规则、领取快照、回写事件，并执行人工处理。
type Usecase struct {
	rules       repository.AlertRuleRepo
	events      repository.AlertEventRepo
	datasources repository.DatasourceRepo
	locks       *lock.Client
	access      access.Checker
	log         *log.Helper
}

func NewUsecase(rules repository.AlertRuleRepo, events repository.AlertEventRepo, datasources repository.DatasourceRepo, locks *lock.Client, checker access.Checker, logger log.Logger) *Usecase {
	return &Usecase{rules: rules, events: events, datasources: datasources, locks: locks, access: checker, log: log.NewHelper(logger)}
}

// Schedule 在锁内把到期规则的下一次评测时间推迟一个周期，并记下本轮批次。拿到锁后不续期。
func (uc *Usecase) Schedule(ctx context.Context) error {
	if uc.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := uc.locks.Acquire(ctx, ScheduleLockKey, EvalInterval)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	now := time.Now().UTC().Truncate(time.Second)
	if err := uc.rules.ScheduleDue(ctx, now, now.Add(EvalInterval)); err != nil {
		uc.log.Errorf("schedule eval: %v", err)
		return err
	}
	return nil
}

// Claim 返回已调度且尚未回写的规则。只接受名为 megrez 的 service key。
func (uc *Usecase) Claim(ctx context.Context) ([]bo.RuleSnapshot, error) {
	if err := uc.requireMegrez(ctx); err != nil {
		return nil, err
	}
	rules, err := uc.rules.ListScheduled(ctx)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	snapshots := make([]bo.RuleSnapshot, 0, len(rules))
	for _, rule := range rules {
		if rule.ScheduledSlot == nil {
			continue
		}
		recorded, err := uc.events.HasBatch(ctx, rule.UID, *rule.ScheduledSlot)
		if err != nil {
			return nil, err
		}
		if recorded {
			continue
		}
		source, err := uc.datasources.GetDatasource(ctx, rule.NamespaceUID, rule.DatasourceUID)
		if err != nil {
			if merr.IsNotFound(err) {
				source = &bo.Datasource{}
			} else {
				return nil, uc.access.Wrap(err)
			}
		}
		fingerprints, err := uc.events.ListFiringFingerprints(ctx, rule.NamespaceUID, rule.UID)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, bo.RuleSnapshot{Rule: rule, Datasource: *source, FiringFingerprints: fingerprints})
	}
	return snapshots, nil
}

// Report 按 fingerprint 幂等回写。重复批次直接返回。
func (uc *Usecase) Report(ctx context.Context, report bo.EvalReport) error {
	if err := uc.requireMegrez(ctx); err != nil {
		return err
	}
	if report.RuleUID == "" || report.EvalSlot.IsZero() {
		return merr.Invalid("invalid request")
	}
	rule, err := uc.rules.GetAlertRuleByUID(ctx, report.RuleUID)
	if err != nil {
		return uc.access.Wrap(err)
	}
	report.RuleUID = rule.UID
	if err := uc.events.ApplyReport(ctx, *rule, report); err != nil {
		return uc.access.Wrap(err)
	}
	return nil
}

func (uc *Usecase) List(ctx context.Context, view, name string, pageNo, pageSize int32) ([]bo.AlertEvent, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if view == "" {
		view = "REALTIME"
	}
	if view != "REALTIME" && view != "HISTORY" {
		return nil, 0, 0, 0, merr.Invalid("invalid request")
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.events.ListEvents(ctx, ns, view, name, p, s)
	if err != nil {
		return nil, 0, 0, 0, uc.access.Wrap(err)
	}
	return items, total, p, s, nil
}

func (uc *Usecase) Get(ctx context.Context, uid string) (*bo.AlertEvent, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.events.GetEvent(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Stats(ctx context.Context) (bo.AlertStats, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return bo.AlertStats{}, err
	}
	stats, err := uc.events.Stats(ctx, ns)
	return stats, uc.access.Wrap(err)
}

func (uc *Usecase) Intervene(ctx context.Context, uids []string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertIntervene)
	if err != nil {
		return err
	}
	userUID, _ := actor.UserID(ctx)
	return uc.access.Wrap(uc.events.Intervene(ctx, ns, uids, userUID))
}

func (uc *Usecase) Suppress(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertIntervene)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.events.Suppress(ctx, ns, uid))
}

func (uc *Usecase) Recover(ctx context.Context, uids []string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertIntervene)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.events.RecoverManual(ctx, ns, uids))
}

func (uc *Usecase) requireMegrez(ctx context.Context) error {
	name, ok := actor.ServiceName(ctx)
	if !ok || name != "megrez" {
		return merr.Forbidden("service key is required")
	}
	return nil
}
