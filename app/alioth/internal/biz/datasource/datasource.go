package datasource

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	datasources  repository.DatasourceRepo
	dictionaries repository.AlertDictionaryRepo
	access       access.Checker
}

func NewUsecase(datasources repository.DatasourceRepo, dictionaries repository.AlertDictionaryRepo, checker access.Checker) *Usecase {
	return &Usecase{datasources: datasources, dictionaries: dictionaries, access: checker}
}

// CreateDatasource 先核对类型和驱动是否成对，再确认数据源级别属于当前租户。
// 密码和密钥由仓储加密，这里只拒绝空地址和错配的驱动。
func (uc *Usecase) CreateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	if err := validateDatasource(in); err != nil {
		return nil, err
	}
	if err := uc.requireDatasourceLevel(ctx, ns, in.LevelUID); err != nil {
		return nil, err
	}
	labels, err := labelsutil.NormalizeObject(in.Labels)
	if err != nil {
		return nil, err
	}
	in.Labels = labels
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.datasources.CreateDatasource(ctx, in)
	return redact(item), uc.access.Wrap(err)
}

func (uc *Usecase) UpdateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	if err := validateDatasource(in); err != nil {
		return nil, err
	}
	if err := uc.requireDatasourceLevel(ctx, ns, in.LevelUID); err != nil {
		return nil, err
	}
	labels, err := labelsutil.NormalizeObject(in.Labels)
	if err != nil {
		return nil, err
	}
	in.Labels = labels
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.datasources.UpdateDatasource(ctx, in)
	return redact(item), uc.access.Wrap(err)
}

func (uc *Usecase) UpdateDatasourceStatus(ctx context.Context, uid, status string) (*bo.Datasource, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.datasources.UpdateDatasourceStatus(ctx, ns, uid, status)
	return redact(item), uc.access.Wrap(err)
}

func (uc *Usecase) DeleteDatasource(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.datasources.DeleteDatasource(ctx, ns, uid))
}

func (uc *Usecase) GetDatasource(ctx context.Context, uid string) (*bo.Datasource, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.datasources.GetDatasource(ctx, ns, uid)
	return redact(item), uc.access.Wrap(err)
}

func (uc *Usecase) ListDatasources(ctx context.Context, pageNo, pageSize int32, name, kind string) ([]bo.Datasource, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.datasources.ListDatasources(ctx, ns, p, s, strings.TrimSpace(name), kind)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) SelectDatasources(ctx context.Context, limit int32, lastUID, kind string) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.datasources.SelectDatasources(ctx, ns, page.Limit(limit), lastUID, kind)
	return items, uc.access.Wrap(err)
}

func (uc *Usecase) Status(ctx context.Context, uid string) (string, error) {
	item, err := uc.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	status, err := uc.datasources.DatasourceStatus(ctx, *item)
	return status, uc.access.Wrap(err)
}

func (uc *Usecase) Metrics(ctx context.Context, uid string) ([]string, error) {
	item, err := uc.metricSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	names, err := uc.datasources.ListMetrics(ctx, *item)
	return names, uc.access.Wrap(err)
}

func (uc *Usecase) Metric(ctx context.Context, uid, name string) (string, error) {
	item, err := uc.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	body, err := uc.datasources.MetricDetail(ctx, *item, name)
	return body, uc.access.Wrap(err)
}

func (uc *Usecase) Query(ctx context.Context, uid, expr string, at int64) (string, error) {
	item, err := uc.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	body, err := uc.datasources.InstantQuery(ctx, *item, expr, at)
	return body, uc.access.Wrap(err)
}

func (uc *Usecase) QueryRange(ctx context.Context, uid, expr string, start, end, step int64) (string, error) {
	item, err := uc.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	if end < start {
		return "", merr.Invalid("invalid request")
	}
	body, err := uc.datasources.RangeQuery(ctx, *item, expr, start, end, step)
	return body, uc.access.Wrap(err)
}

// Proxy 把查询原样转到数据源。日志和链路只走这里，不写评测批次或告警事件。
func (uc *Usecase) Proxy(ctx context.Context, uid, method, path, body string) (string, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return "", err
	}
	item, err := uc.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return "", uc.access.Wrap(err)
	}
	if item.Status != "ACTIVE" {
		return "", merr.Invalid("datasource is disabled")
	}
	result, err := uc.datasources.ProxyQuery(ctx, *item, method, path, body)
	return result, uc.access.Wrap(err)
}

func (uc *Usecase) metricSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if item.Type != "METRICS" {
		return nil, merr.Invalid("datasource is not metrics")
	}
	if item.Status != "ACTIVE" {
		return nil, merr.Invalid("datasource is disabled")
	}
	return item, nil
}

func (uc *Usecase) requireDatasourceLevel(ctx context.Context, namespaceUID, levelUID string) error {
	if levelUID == "" {
		return nil
	}
	level, err := uc.dictionaries.GetAlertDictionary(ctx, namespaceUID, levelUID)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if level.Type != "DATASOURCE" || level.Status != "ACTIVE" {
		return merr.Invalid("invalid request")
	}
	return nil
}

func validateDatasource(in bo.Datasource) error {
	if !driverMatches(in.Type, in.Driver) || strings.TrimSpace(in.Endpoint) == "" {
		return merr.Invalid("invalid request")
	}
	return nil
}

func driverMatches(kind, driver string) bool {
	switch kind {
	case "METRICS":
		return driver == "METRICS_PROMETHEUS" || driver == "METRICS_VICTORIA_METRICS"
	case "LOGS":
		return driver == "LOGS_ELASTICSEARCH"
	case "TRACE":
		return driver == "TRACE_JAEGER"
	default:
		return false
	}
}

func redact(item *bo.Datasource) *bo.Datasource {
	if item == nil {
		return nil
	}
	item.Password = ""
	item.Secret = ""
	return item
}
