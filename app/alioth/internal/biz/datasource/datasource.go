package datasource

import (
	"context"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/endpoint"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// defaultLogLimit 是 Loki 日志查询未指定条数时的上限。
	defaultLogLimit = 100
	// maxLogLimit 是单次日志查询允许返回的最大条数。
	maxLogLimit = 5000
	// maxDatasourceName 与 proto 校验保持一致，避免绕过 HTTP 的调用写入超长名称。
	maxDatasourceName = 64
)

// Gateway 管理数据源配置，并把查询请求转发给对应驱动。
type Gateway struct {
	datasources  repository.DatasourceRepo
	dictionaries repository.AlertDictionaryRepo
	events       repository.AlertEventRepo
	access       access.Checker
}

func NewGateway(datasources repository.DatasourceRepo, dictionaries repository.AlertDictionaryRepo, events repository.AlertEventRepo, checker access.Checker) *Gateway {
	return &Gateway{datasources: datasources, dictionaries: dictionaries, events: events, access: checker}
}

// CreateDatasource 先核对类型和驱动是否成对，再确认数据源级别属于当前租户。
// 密码和密钥由仓储加密，这里只拒绝空地址和错配的驱动。
func (gateway *Gateway) CreateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	normalizeDatasource(&in)
	if err := validateDatasource(in); err != nil {
		return nil, err
	}
	if err := gateway.requireDatasourceLevel(ctx, ns, in.LevelUID); err != nil {
		return nil, err
	}
	labels, err := labelsutil.NormalizeObject(in.Labels)
	if err != nil {
		return nil, err
	}
	in.Labels = labels
	in.NamespaceUID = ns
	item, err := gateway.datasources.CreateDatasource(ctx, in)
	return redact(item), gateway.access.Wrap(err)
}

func (gateway *Gateway) UpdateDatasource(ctx context.Context, in bo.Datasource) (*bo.Datasource, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	normalizeDatasource(&in)
	if err := validateDatasource(in); err != nil {
		return nil, err
	}
	if err := gateway.requireDatasourceLevel(ctx, ns, in.LevelUID); err != nil {
		return nil, err
	}
	labels, err := labelsutil.NormalizeObject(in.Labels)
	if err != nil {
		return nil, err
	}
	in.Labels = labels
	in.NamespaceUID = ns
	previous, err := gateway.datasources.GetDatasource(ctx, ns, in.UID)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	item, err := gateway.datasources.UpdateDatasource(ctx, in)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	// 等级、标签、类型或驱动变了，旧事件要么级别名称过期，要么可能不再被规则匹配。
	// 先收口成恢复，下一轮评测会按新配置重新产生事件；否则旧事件会一直挂在实时列表，
	// 用户看到的就是「规则已经满足条件，但没有触发新事件」。
	if previous.LevelUID != item.LevelUID || previous.Labels != item.Labels ||
		previous.Type != item.Type || previous.Driver != item.Driver {
		if err := gateway.events.RecoverSource(ctx, ns, nil, []string{item.UID}, nil); err != nil {
			return nil, gateway.access.Wrap(err)
		}
	}
	return redact(item), nil
}

func (gateway *Gateway) UpdateDatasourceStatus(ctx context.Context, uid, status string) (*bo.Datasource, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceWrite)
	if err != nil {
		return nil, err
	}
	item, err := gateway.datasources.UpdateDatasourceStatus(ctx, ns, uid, status)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if status == "DISABLED" {
		if err := gateway.events.RecoverSource(ctx, ns, nil, []string{uid}, nil); err != nil {
			return nil, gateway.access.Wrap(err)
		}
	}
	return redact(item), nil
}

func (gateway *Gateway) DeleteDatasource(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceDelete)
	if err != nil {
		return err
	}
	if err := gateway.events.RecoverSource(ctx, ns, nil, []string{uid}, nil); err != nil {
		return gateway.access.Wrap(err)
	}
	return gateway.access.Wrap(gateway.datasources.DeleteDatasource(ctx, ns, uid))
}

func (gateway *Gateway) GetDatasource(ctx context.Context, uid string) (*bo.Datasource, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.datasources.GetDatasource(ctx, ns, uid)
	return redact(item), gateway.access.Wrap(err)
}

func (gateway *Gateway) ListDatasources(ctx context.Context, pageNo, pageSize int32, name, kind string) ([]bo.Datasource, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := gateway.datasources.ListDatasources(ctx, ns, p, s, strings.TrimSpace(name), kind)
	return items, total, p, s, gateway.access.Wrap(err)
}

func (gateway *Gateway) SelectDatasources(ctx context.Context, limit int32, lastUID, kind string) ([]bo.SelectItem, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	items, err := gateway.datasources.SelectDatasources(ctx, ns, page.Limit(limit), lastUID, kind)
	return items, gateway.access.Wrap(err)
}

func (gateway *Gateway) Status(ctx context.Context, uid string) (string, error) {
	item, err := gateway.statusSource(ctx, uid)
	if err != nil {
		return "", err
	}
	status, err := gateway.datasources.DatasourceStatus(ctx, *item)
	return status, gateway.access.Wrap(err)
}

func (gateway *Gateway) Metrics(ctx context.Context, uid string) ([]bo.DatasourceMetric, error) {
	item, err := gateway.metricSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	metrics, err := gateway.datasources.ListMetrics(ctx, *item)
	return metrics, gateway.access.Wrap(err)
}

// Metadata 读取日志数据源的关键信息。指标数据源继续走 Metrics，
// 这条路径只服务 Loki、VictoriaLogs 和 Elasticsearch。
func (gateway *Gateway) Metadata(ctx context.Context, uid string) (*bo.DatasourceMetadata, error) {
	item, err := gateway.logSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	metadata, err := gateway.datasources.DatasourceMetadata(ctx, *item)
	return metadata, gateway.access.Wrap(err)
}

func (gateway *Gateway) Metric(ctx context.Context, uid, name string) (string, error) {
	item, err := gateway.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	body, err := gateway.datasources.MetricDetail(ctx, *item, name)
	return body, gateway.access.Wrap(err)
}

func (gateway *Gateway) Query(ctx context.Context, uid, expr string, at int64) (string, error) {
	item, err := gateway.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	body, err := gateway.datasources.InstantQuery(ctx, *item, expr, at)
	return body, gateway.access.Wrap(err)
}

func (gateway *Gateway) QueryRange(ctx context.Context, uid, expr string, start, end, step int64) (string, error) {
	item, err := gateway.metricSource(ctx, uid)
	if err != nil {
		return "", err
	}
	if end < start {
		return "", merr.Invalid("invalid request")
	}
	body, err := gateway.datasources.RangeQuery(ctx, *item, expr, start, end, step)
	return body, gateway.access.Wrap(err)
}

// QueryLogs 查询日志数据源。Loki 使用 LogQL，VictoriaLogs 使用 LogsQL，
// Elasticsearch 使用 Query DSL 或普通检索语句。时间窗口、条数和方向在这里收敛，
// 避免把任意参数透传给上游。
func (gateway *Gateway) QueryLogs(ctx context.Context, uid, query string, start, end int64, limit int32, direction string) (string, error) {
	item, err := gateway.logSource(ctx, uid)
	if err != nil {
		return "", err
	}
	query, size, dir, err := normalizeLogQuery(query, start, end, limit, direction)
	if err != nil {
		return "", err
	}
	body, err := gateway.datasources.QueryLogs(ctx, *item, query, start, end, size, dir)
	return body, gateway.access.Wrap(err)
}

// Proxy 把查询原样转到数据源。未单独建模的日志和链路查询走这里，不写评测批次或告警事件。
func (gateway *Gateway) Proxy(ctx context.Context, uid, method, path, body string) (string, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return "", err
	}
	item, err := gateway.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return "", gateway.access.Wrap(err)
	}
	if item.Status != "ACTIVE" {
		return "", merr.Invalid("datasource is disabled")
	}
	result, err := gateway.datasources.ProxyQuery(ctx, *item, method, path, body)
	return result, gateway.access.Wrap(err)
}

func (gateway *Gateway) metricSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	item, err := gateway.readSource(ctx, uid)
	if err != nil {
		return nil, err
	}
	if item.Type != "METRICS" {
		return nil, merr.Invalid("datasource is not metrics")
	}
	return item, nil
}

// logSource 取日志数据源。Loki、VictoriaLogs 与 Elasticsearch 都支持日志查询。
func (gateway *Gateway) logSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	item, err := gateway.readSource(ctx, uid)
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
func (gateway *Gateway) statusSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	item, err := gateway.readSource(ctx, uid)
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
func (gateway *Gateway) readSource(ctx context.Context, uid string) (*bo.Datasource, error) {
	ns, err := gateway.access.Allow(ctx, perm.DatasourceRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if item.Status != "ACTIVE" {
		return nil, merr.Invalid("datasource is disabled")
	}
	return item, nil
}

func (gateway *Gateway) requireDatasourceLevel(ctx context.Context, namespaceUID, levelUID string) error {
	if levelUID == "" {
		return nil
	}
	level, err := gateway.dictionaries.GetAlertDictionary(ctx, namespaceUID, levelUID)
	if err != nil {
		return gateway.access.Wrap(err)
	}
	if level.Type != "DATASOURCE" || level.Status != "ACTIVE" {
		return merr.Invalid("invalid request")
	}
	return nil
}

func validateDatasource(in bo.Datasource) error {
	if !DriverMatches(in.Type, in.Driver) {
		return merr.Invalid("invalid request")
	}
	if in.Name == "" {
		return merr.Invalid("invalid request")
	}
	if utf8.RuneCountInString(in.Name) > maxDatasourceName {
		return merr.Invalid("name is too long")
	}
	// 地址必须能被数据层拼成上游请求。`htp://` 这类拼写错误如果放进来，
	// 要等到用户点查询才在数据层报 invalid request，保存时就该拒绝。
	endpoint, err := url.Parse(strings.TrimSpace(in.Endpoint))
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return merr.Invalid("invalid request")
	}
	return nil
}

// normalizeDatasource 清理名称和地址，并清理只在 Elasticsearch 上生效的设置，
// 避免切换驱动后留下旧值。地址去掉末尾斜杠后再做唯一性判断。
func normalizeDatasource(in *bo.Datasource) {
	in.Name = strings.TrimSpace(in.Name)
	in.Endpoint = endpoint.Normalize(in.Endpoint)
	in.IndexPattern = strings.TrimSpace(in.IndexPattern)
	in.TimeField = strings.TrimSpace(in.TimeField)
	if in.Driver != "LOGS_ELASTICSEARCH" {
		in.IndexPattern = ""
		in.TimeField = ""
	}
}

// DriverMatches 判断数据源类型和驱动是否匹配。
func DriverMatches(kind, driver string) bool {
	return TypeForDriver(driver) == kind
}

// TypeForDriver 从驱动推导数据源类型。告警规则只保存驱动，类型由这里统一推导。
func TypeForDriver(driver string) string {
	switch driver {
	case "METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS":
		return "METRICS"
	case "LOGS_ELASTICSEARCH", "LOGS_LOKI", "LOGS_VLOGS":
		return "LOGS"
	case "TRACE_JAEGER":
		return "TRACE"
	default:
		return ""
	}
}

// ValidDriver 判断驱动是否属于已支持的枚举。
func ValidDriver(driver string) bool {
	return TypeForDriver(driver) != ""
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
