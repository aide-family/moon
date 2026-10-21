package alertdictionary

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	dictionaries repository.AlertDictionaryRepo
	datasources  repository.DatasourceRepo
	events       repository.AlertEventRepo
	access       access.Checker
}

func NewUsecase(dictionaries repository.AlertDictionaryRepo, datasources repository.DatasourceRepo, events repository.AlertEventRepo, checker access.Checker) *Usecase {
	return &Usecase{dictionaries: dictionaries, datasources: datasources, events: events, access: checker}
}

func (uc *Usecase) CreateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.dictionaries.CreateAlertDictionary(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	previous, err := uc.dictionaries.GetAlertDictionary(ctx, ns, in.UID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	item, err := uc.dictionaries.UpdateAlertDictionary(ctx, in)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	// datasource_level 匹配器读的是等级名称。等级改名后匹配集合会变化，
	// 先恢复使用该等级的数据源事件，下一轮评测按新名称重新匹配。
	if previous.Type == "DATASOURCE" && previous.Name != item.Name {
		if err := uc.recoverLevelDatasources(ctx, ns, item.UID); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func (uc *Usecase) UpdateAlertDictionaryStatus(ctx context.Context, uid, status string) (*bo.AlertDictionary, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.dictionaries.UpdateAlertDictionaryStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

// DeleteAlertDictionary 拒绝删除仍挂在规则上的字典项，避免条件指向一条已经不存在的级别。
func (uc *Usecase) DeleteAlertDictionary(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return err
	}
	item, err := uc.dictionaries.GetAlertDictionary(ctx, ns, uid)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if item.Type == "DATASOURCE" {
		if err := uc.recoverLevelDatasources(ctx, ns, uid); err != nil {
			return err
		}
	}
	return uc.access.Wrap(uc.dictionaries.DeleteAlertDictionary(ctx, ns, uid))
}

// recoverLevelDatasources 收口使用某个数据源等级的所有事件。
func (uc *Usecase) recoverLevelDatasources(ctx context.Context, namespaceUID, levelUID string) error {
	datasourceUIDs, err := uc.datasources.ListDatasourceUIDsByLevel(ctx, namespaceUID, levelUID)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if len(datasourceUIDs) == 0 {
		return nil
	}
	return uc.access.Wrap(uc.events.RecoverSource(ctx, namespaceUID, nil, datasourceUIDs, nil))
}

func (uc *Usecase) ListAlertDictionaries(ctx context.Context, pageNo, pageSize int32, kind, name string) ([]bo.AlertDictionary, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.dictionaries.ListAlertDictionaries(ctx, ns, p, s, kind, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}
