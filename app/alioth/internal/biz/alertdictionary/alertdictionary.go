package alertdictionary

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

var ProviderSet = wire.NewSet(NewCatalog)

// Catalog 维护告警字典（级别与分类），并保持数据源引用一致。
type Catalog struct {
	dictionaries repository.AlertDictionaryRepo
	datasources  repository.DatasourceRepo
	events       repository.AlertEventRepo
	access       access.Checker
}

func NewCatalog(dictionaries repository.AlertDictionaryRepo, datasources repository.DatasourceRepo, events repository.AlertEventRepo, checker access.Checker) *Catalog {
	return &Catalog{dictionaries: dictionaries, datasources: datasources, events: events, access: checker}
}

func (catalog *Catalog) CreateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	ns, err := catalog.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := catalog.dictionaries.CreateAlertDictionary(ctx, in)
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) UpdateAlertDictionary(ctx context.Context, in bo.AlertDictionary) (*bo.AlertDictionary, error) {
	ns, err := catalog.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	previous, err := catalog.dictionaries.GetAlertDictionary(ctx, ns, in.UID)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	item, err := catalog.dictionaries.UpdateAlertDictionary(ctx, in)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	// datasource_level 匹配器读的是等级名称。等级改名后匹配集合会变化，
	// 先恢复使用该等级的数据源事件，下一轮评测按新名称重新匹配。
	if previous.Type == "DATASOURCE" && previous.Name != item.Name {
		if err := catalog.recoverLevelDatasources(ctx, ns, item.UID); err != nil {
			return nil, err
		}
	}
	return item, nil
}

func (catalog *Catalog) UpdateAlertDictionaryStatus(ctx context.Context, uid, status string) (*bo.AlertDictionary, error) {
	ns, err := catalog.access.Allow(ctx, perm.AlertRuleWrite)
	if err != nil {
		return nil, err
	}
	item, err := catalog.dictionaries.UpdateAlertDictionaryStatus(ctx, ns, uid, status)
	return item, catalog.access.Wrap(err)
}

// DeleteAlertDictionary 拒绝删除仍挂在规则上的字典项，避免条件指向一条已经不存在的级别。
func (catalog *Catalog) DeleteAlertDictionary(ctx context.Context, uid string) error {
	ns, err := catalog.access.Allow(ctx, perm.AlertRuleDelete)
	if err != nil {
		return err
	}
	item, err := catalog.dictionaries.GetAlertDictionary(ctx, ns, uid)
	if err != nil {
		return catalog.access.Wrap(err)
	}
	if item.Type == "DATASOURCE" {
		if err := catalog.recoverLevelDatasources(ctx, ns, uid); err != nil {
			return err
		}
	}
	return catalog.access.Wrap(catalog.dictionaries.DeleteAlertDictionary(ctx, ns, uid))
}

// recoverLevelDatasources 收口使用某个数据源等级的所有事件。
func (catalog *Catalog) recoverLevelDatasources(ctx context.Context, namespaceUID, levelUID string) error {
	datasourceUIDs, err := catalog.datasources.ListDatasourceUIDsByLevel(ctx, namespaceUID, levelUID)
	if err != nil {
		return catalog.access.Wrap(err)
	}
	if len(datasourceUIDs) == 0 {
		return nil
	}
	return catalog.access.Wrap(catalog.events.RecoverSource(ctx, namespaceUID, nil, datasourceUIDs, nil))
}

func (catalog *Catalog) ListAlertDictionaries(ctx context.Context, pageNo, pageSize int32, kind, name string) ([]bo.AlertDictionary, int, int, int, error) {
	ns, err := catalog.access.Allow(ctx, perm.AlertRuleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := catalog.dictionaries.ListAlertDictionaries(ctx, ns, p, s, kind, strings.TrimSpace(name))
	return items, total, p, s, catalog.access.Wrap(err)
}
