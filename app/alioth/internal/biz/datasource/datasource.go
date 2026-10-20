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

const (
	// defaultLogLimit 是 Loki 日志查询未指定条数时的上限。
	defaultLogLimit = 100
	// maxLogLimit 是单次日志查询允许返回的最大条数。
	maxLogLimit = 5000
)

type Usecase struct {
	datasources  repository.DatasourceRepo
	dictionaries repository.AlertDictionaryRepo
	events       repository.AlertEventRepo
	access       access.Checker
}

func NewUsecase(datasources repository.DatasourceRepo, dictionaries repository.AlertDictionaryRepo, events repository.AlertEventRepo, checker access.Checker) *Usecase {
	return &Usecase{datasources: datasources, dictionaries: dictionaries, events: events, access: checker}
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
	normalizeDatasource(&in)
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
	normalizeDatasource(&in)
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
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if status == "DISABLED" {
		if err := uc.events.RecoverSource(ctx, ns, nil, []string{uid}, nil); err != nil {
			return nil, uc.access.Wrap(err)
		}
	}
	return redact(item), nil
}

func (uc *Usecase) DeleteDatasource(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return err
	}
	if err := uc.events.RecoverSource(ctx, ns, nil, []string{uid}, nil); err != nil {
		return uc.access.Wrap(err)
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
	item, err := uc.statusSource(ctx, uid)
	if err != nil {
		return "", err
	}
	status, err := uc.datasources.DatasourceStatus(ctx, *item)
	return status, uc.access.Wrap(err)
}

func (uc *Usecase) Metrics(ctx context.Context, uid string) ([]bo.DatasourceMetric, error) {
	item, err := uc.metricSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	metrics, err := uc.datasources.ListMetrics(ctx, *item)
	return metrics, uc.access.Wrap(err)
}

// Metadata 读取日志数据源的关键信息。指标数据源继续走 Metrics，
// 这条路径只服务 Loki、VictoriaLogs 和 Elasticsearch。
func (uc *Usecase) Metadata(ctx context.Context, uid string) (*bo.DatasourceMetadata, error) {
	item, err := uc.logSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	metadata, err := uc.datasources.DatasourceMetadata(ctx, *item)
	return metadata, uc.access.Wrap(err)
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

// QueryLogs 查询日志数据源。Loki 使用 LogQL，VictoriaLogs 使用 LogsQL，
// Elasticsearch 使用 Query DSL 或普通检索语句。时间窗口、条数和方向在这里收敛，
// 避免把任意参数透传给上游。
func (uc *Usecase) QueryLogs(ctx context.Context, uid, query string, start, end int64, limit int32, direction string) (string, error) {
	item, err := uc.logSource(ctx, uid)
	if err != nil {
		return "", err
	}
	query, size, dir, err := normalizeLogQuery(query, start, end, limit, direction)
	if err != nil {
		return "", err
	}
	body, err := uc.datasources.QueryLogs(ctx, *item, query, start, end, size, dir)
	return body, uc.access.Wrap(err)
}

// Proxy 把查询原样转到数据源。未单独建模的日志和链路查询走这里，不写评测批次或告警事件。
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
	item, err := uc.readSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	if item.Type != "METRICS" {
		return nil, merr.Invalid("datasource is not metrics")
	}
	return item, nil
}

// logSource 取日志数据源。Loki、VictoriaLogs 与 Elasticsearch 都支持日志查询。
func (uc *Usecase) logSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	item, err := uc.readSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	if item.Type != "LOGS" {
		return nil, merr.Invalid("datasource is not logs")
	}
	switch item.Driver {
	case "LOGS_LOKI", "LOGS_VLOGS", "LOGS_ELASTICSEARCH":
		return item, nil
	default:
		return nil, merr.Invalid("invalid request")
	}
}

// statusSource 允许指标数据源、Loki、VictoriaLogs 和 Elasticsearch 日志数据源做连通检查。
func (uc *Usecase) statusSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	item, err := uc.readSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	switch item.Driver {
	case "METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS", "LOGS_LOKI", "LOGS_VLOGS", "LOGS_ELASTICSEARCH":
		return item, nil
	default:
		return nil, merr.Invalid("invalid request")
	}
}

// readSource 读取当前租户的数据源并确认已启用。类型和驱动限制由调用方补充。
func (uc *Usecase) readSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	ns, err := uc.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
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
	if !DriverMatches(in.Type, in.Driver) || strings.TrimSpace(in.Endpoint) == "" {
		return merr.Invalid("invalid request")
	}
	return nil
}

// normalizeDatasource 清理只在 Elasticsearch 上生效的设置，避免切换驱动后留下旧值。
func normalizeDatasource(in *bo.Datasource) {
	in.IndexPattern = strings.TrimSpace(in.IndexPattern)
	in.TimeField = strings.TrimSpace(in.TimeField)
	if in.Driver != "LOGS_ELASTICSEARCH" {
		in.IndexPattern = ""
		in.TimeField = ""
	}
}

// DriverMatches 判断数据源类型和驱动是否匹配。
func DriverMatches(kind, driver string) bool {
	switch kind {
	case "METRICS":
		return driver == "METRICS_PROMETHEUS" || driver == "METRICS_VICTORIA_METRICS"
	case "LOGS":
		return driver == "LOGS_ELASTICSEARCH" || driver == "LOGS_LOKI" || driver == "LOGS_VLOGS"
	case "TRACE":
		return driver == "TRACE_JAEGER"
	default:
		return false
	}
}

// normalizeLogQuery 校验日志查询参数。limit 为 0 时取默认值，超过上限按上限截断；
// direction 留空按默认的 backward 处理，日志驱动各自映射为对应的排序方向。
func normalizeLogQuery(query string, start, end int64, limit int32, direction string) (string, int, string, error) {
	query = strings.TrimSpace(query)
	if query == "" || start < 0 || end <= start {
		return "", 0, "", merr.Invalid("invalid request")
	}
	size := int(limit)
	if size <= 0 {
		size = defaultLogLimit
	}
	if size > maxLogLimit {
		size = maxLogLimit
	}
	switch direction {
	case "":
		direction = "backward"
	case "backward", "forward":
	default:
		return "", 0, "", merr.Invalid("invalid request")
	}
	return query, size, direction, nil
}

func redact(item *bo.Datasource) *bo.Datasource {
	if item == nil {
		return nil
	}
	item.Password = ""
	item.Secret = ""
	return item
}
