package datasource

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func TestRedactClearsSecrets(t *testing.T) {
	item := redact(&bo.Datasource{Password: "plain-password", Secret: "plain-secret", SecretSet: true})
	if item.Password != "" || item.Secret != "" || !item.SecretSet {
		t.Fatalf("redacted = %+v", item)
	}
}

func TestDriverMatchesLoki(t *testing.T) {
	if !DriverMatches("LOGS", "LOGS_LOKI") {
		t.Fatal("LOGS/LOGS_LOKI should match")
	}
	if !DriverMatches("LOGS", "LOGS_VLOGS") {
		t.Fatal("LOGS/LOGS_VLOGS should match")
	}
	if DriverMatches("METRICS", "LOGS_LOKI") {
		t.Fatal("METRICS/LOGS_LOKI should not match")
	}
	if DriverMatches("LOGS", "LOGS_UNKNOWN") {
		t.Fatal("unknown logs driver should not match")
	}
}

func TestValidateDatasourceEndpoint(t *testing.T) {
	valid := []string{
		"http://127.0.0.1:9090",
		"https://metrics.example.com/prometheus",
		" http://victoriametrics.moon.com/select/0/prometheus ",
	}
	for _, endpoint := range valid {
		if err := validateDatasource(bo.Datasource{Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: endpoint}); err != nil {
			t.Fatalf("validateDatasource(%q) error = %v", endpoint, err)
		}
	}

	invalid := []string{
		"",
		"htp://es.moon.com",
		"es.moon.com",
		"ftp://es.moon.com",
		"http://",
		"http://exa mple.com",
	}
	for _, endpoint := range invalid {
		if err := validateDatasource(bo.Datasource{Name: "es", Type: "LOGS", Driver: "LOGS_ELASTICSEARCH", Endpoint: endpoint}); !merr.IsInvalid(err) {
			t.Fatalf("validateDatasource(%q) error = %v", endpoint, err)
		}
	}

	if err := validateDatasource(bo.Datasource{Name: "bad", Type: "METRICS", Driver: "LOGS_LOKI", Endpoint: "http://loki"}); !merr.IsInvalid(err) {
		t.Fatalf("mismatched driver error = %v", err)
	}
	if err := validateDatasource(bo.Datasource{Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: "http://prom"}); !merr.IsInvalid(err) {
		t.Fatalf("empty name error = %v", err)
	}
	if err := validateDatasource(bo.Datasource{Name: strings.Repeat("长", maxDatasourceName+1), Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: "http://prom"}); !merr.IsInvalid(err) {
		t.Fatalf("long name error = %v", err)
	}
}

func TestNormalizeDatasourceEndpoint(t *testing.T) {
	item := bo.Datasource{Name: " prom ", Endpoint: " HTTP://Prometheus.Moon.com:9090/ "}
	normalizeDatasource(&item)
	if item.Name != "prom" || item.Endpoint != "http://prometheus.moon.com:9090" {
		t.Fatalf("normalized = %+v", item)
	}
}

func TestNormalizeDatasourceKeepsElasticsearchSettings(t *testing.T) {
	item := bo.Datasource{Driver: "LOGS_ELASTICSEARCH", IndexPattern: " logs-* ", TimeField: " event.time "}
	normalizeDatasource(&item)
	if item.IndexPattern != "logs-*" || item.TimeField != "event.time" {
		t.Fatalf("normalized = %+v", item)
	}

	item = bo.Datasource{Driver: "LOGS_LOKI", IndexPattern: "logs-*", TimeField: "event.time"}
	normalizeDatasource(&item)
	if item.IndexPattern != "" || item.TimeField != "" {
		t.Fatalf("loki settings should be cleared: %+v", item)
	}
}

func TestNormalizeLogQuery(t *testing.T) {
	query, size, direction, err := normalizeLogQuery(`{app="api"} |= "error"`, 100, 200, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if query != `{app="api"} |= "error"` || size != defaultLogLimit || direction != "backward" {
		t.Fatalf("normalized = %q %d %q", query, size, direction)
	}

	_, size, direction, err = normalizeLogQuery("{}", 1, 2, maxLogLimit+1, "forward")
	if err != nil {
		t.Fatal(err)
	}
	if size != maxLogLimit || direction != "forward" {
		t.Fatalf("normalized size = %d direction = %q", size, direction)
	}

	cases := []struct {
		name      string
		query     string
		start     int64
		end       int64
		limit     int32
		direction string
	}{
		{name: "empty query", query: "  ", start: 1, end: 2},
		{name: "empty window", query: "{}", start: 2, end: 2},
		{name: "reversed window", query: "{}", start: 3, end: 2},
		{name: "negative start", query: "{}", start: -1, end: 2},
		{name: "unknown direction", query: "{}", start: 1, end: 2, direction: "sideways"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := normalizeLogQuery(tc.query, tc.start, tc.end, tc.limit, tc.direction); !merr.IsInvalid(err) {
				t.Fatalf("normalizeLogQuery error = %v", err)
			}
		})
	}
}

// availabilityDatasource 生成一条地址和驱动都合法、只用于测试可用性字段的数据源。
func availabilityDatasource() bo.Datasource {
	return bo.Datasource{
		Name: "prom", Type: "METRICS", Driver: "METRICS_PROMETHEUS", Endpoint: "http://127.0.0.1:9090",
	}
}

func TestApplyDatasourceAvailabilityDefaults(t *testing.T) {
	item := availabilityDatasource()
	applyDatasourceDefaults(&item)
	if item.FailureWindowMinutes != defaultFailureWindowMinutes || item.FailureThreshold != defaultFailureThreshold {
		t.Fatalf("defaults = %d/%d", item.FailureWindowMinutes, item.FailureThreshold)
	}
	if err := validateDatasource(item); err != nil {
		t.Fatalf("disabled defaults should validate: %v", err)
	}
}

// TestNormalizeDatasourceKeepsAvailabilityZeros 防止默认值在 validate 之前把显式 0 藏起来，
// 否则「启用时必须为正」的校验永远不会失败。
func TestNormalizeDatasourceKeepsAvailabilityZeros(t *testing.T) {
	item := availabilityDatasource()
	item.AvailabilityAlertEnabled = true
	normalizeDatasource(&item)
	if item.FailureWindowMinutes != 0 || item.FailureThreshold != 0 {
		t.Fatalf("normalize filled defaults too early: %d/%d", item.FailureWindowMinutes, item.FailureThreshold)
	}
	if err := validateDatasource(item); !merr.IsInvalid(err) {
		t.Fatalf("enabled with zero values should be invalid, error = %v", err)
	}
}

func TestNormalizeDatasourceNotificationGroups(t *testing.T) {
	item := availabilityDatasource()
	item.NotificationGroupUIDs = []string{" group-1 ", "", "group-2", "group-1"}
	normalizeDatasource(&item)
	if len(item.NotificationGroupUIDs) != 2 || item.NotificationGroupUIDs[0] != "group-1" || item.NotificationGroupUIDs[1] != "group-2" {
		t.Fatalf("group uids = %#v", item.NotificationGroupUIDs)
	}

	empty := availabilityDatasource()
	normalizeDatasource(&empty)
	if empty.NotificationGroupUIDs == nil || len(empty.NotificationGroupUIDs) != 0 {
		t.Fatalf("group uids = %#v, want empty slice", empty.NotificationGroupUIDs)
	}
}

func TestValidateDatasourceAvailability(t *testing.T) {
	valid := []bo.Datasource{
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 1, FailureThreshold: 1},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: maxFailureWindowMinutes, FailureThreshold: maxFailureThreshold},
		{AvailabilityAlertEnabled: false, FailureWindowMinutes: 0, FailureThreshold: 0},
	}
	for _, tc := range valid {
		item := availabilityDatasource()
		item.AvailabilityAlertEnabled = tc.AvailabilityAlertEnabled
		item.FailureWindowMinutes = tc.FailureWindowMinutes
		item.FailureThreshold = tc.FailureThreshold
		if err := validateDatasource(item); err != nil {
			t.Fatalf("validate(%+v) error = %v", tc, err)
		}
	}

	invalid := []bo.Datasource{
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 0, FailureThreshold: 3},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 5, FailureThreshold: 0},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: maxFailureWindowMinutes + 1, FailureThreshold: 3},
		{AvailabilityAlertEnabled: true, FailureWindowMinutes: 5, FailureThreshold: maxFailureThreshold + 1},
		{AvailabilityAlertEnabled: false, FailureWindowMinutes: -1, FailureThreshold: 3},
	}
	for _, tc := range invalid {
		item := availabilityDatasource()
		item.AvailabilityAlertEnabled = tc.AvailabilityAlertEnabled
		item.FailureWindowMinutes = tc.FailureWindowMinutes
		item.FailureThreshold = tc.FailureThreshold
		if err := validateDatasource(item); !merr.IsInvalid(err) {
			t.Fatalf("validate(%+v) error = %v", tc, err)
		}
	}

	tooMany := availabilityDatasource()
	tooMany.NotificationGroupUIDs = make([]string, maxNotificationGroups+1)
	if err := validateDatasource(tooMany); !merr.IsInvalid(err) {
		t.Fatalf("too many groups error = %v", err)
	}
}

// fixNamespaceRepo 只提供当前的启用租户，并让测试操作者成为负责人：
// 负责人隐式持有全部租户权限，下面针对 Create/Update 的用例不需要再铺角色数据。
type fixNamespaceRepo struct{ repository.NamespaceRepo }

func (fixNamespaceRepo) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	return &bo.Namespace{UID: uid, Name: uid, OwnerUID: "fix-owner", Status: "ACTIVE"}, nil
}

// fixFeatureRepo 返回完整的功能与能力集合，让 requireDriverCapability 能通过。
type fixFeatureRepo struct{ repository.FeatureRepo }

func (fixFeatureRepo) EffectiveModules(context.Context, string) ([]string, error) {
	return perm.FeatureCodes(), nil
}

func (fixFeatureRepo) EffectiveCapabilities(context.Context, string) ([]string, error) {
	items := perm.Capabilities()
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	return codes, nil
}

// fixDatasourceRepo 记录用例走到仓储时的入参；existing 模拟修改前已存在的数据源。
type fixDatasourceRepo struct {
	repository.DatasourceRepo
	created  *bo.Datasource
	updated  *bo.Datasource
	existing *bo.Datasource
}

func (r *fixDatasourceRepo) CreateDatasource(_ context.Context, in bo.Datasource) (*bo.Datasource, error) {
	r.created = &in
	item := in
	return &item, nil
}

func (r *fixDatasourceRepo) UpdateDatasource(_ context.Context, in bo.Datasource) (*bo.Datasource, error) {
	r.updated = &in
	item := in
	return &item, nil
}

func (r *fixDatasourceRepo) GetDatasource(context.Context, string, string) (*bo.Datasource, error) {
	if r.existing == nil {
		return nil, merr.NotFound("not found")
	}
	return r.existing, nil
}

func fixDatasourceGateway(sources repository.DatasourceRepo) *Gateway {
	checker := access.Checker{
		Namespaces: fixNamespaceRepo{},
		Features:   fixFeatureRepo{},
		Log:        slog.Default(),
	}
	return NewGateway(sources, nil, nil, nil, checker)
}

// fixDatasourceContext 提供已认证的租户负责人，避免测试依赖角色数据。
func fixDatasourceContext() context.Context {
	return actor.WithUser(actor.WithNamespace(context.Background(), "fix-ns"), "fix-owner")
}

// fixAvailabilityDatasource 生成一条只用于校验可用性告警字段的合法数据源。
func fixAvailabilityDatasource(enabled bool, window, threshold int) bo.Datasource {
	item := availabilityDatasource()
	item.UID = "fix-ds"
	item.Labels = "{}"
	item.AvailabilityAlertEnabled = enabled
	item.FailureWindowMinutes = window
	item.FailureThreshold = threshold
	return item
}

// TestCreateDatasourceRejectsZeroAvailabilityWhenEnabled 复现 M17：显式启用可用性告警时
// 窗口或阈值为 0 必须返回 invalid request，不能先被默认值掩盖。
func TestCreateDatasourceRejectsZeroAvailabilityWhenEnabled(t *testing.T) {
	cases := []struct {
		name      string
		window    int
		threshold int
	}{
		{name: "zero window", window: 0, threshold: 3},
		{name: "zero threshold", window: 5, threshold: 0},
		{name: "zero both", window: 0, threshold: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sources := &fixDatasourceRepo{}
			gateway := fixDatasourceGateway(sources)
			if _, err := gateway.CreateDatasource(fixDatasourceContext(), fixAvailabilityDatasource(true, tc.window, tc.threshold)); !merr.IsInvalid(err) {
				t.Fatalf("CreateDatasource error = %v, want invalid", err)
			}
			if sources.created != nil {
				t.Fatalf("repo should not be called, got %+v", sources.created)
			}
		})
	}
}

// TestCreateDatasourceFillsAvailabilityDefaultsWhenDisabled 校验关闭可用性告警时
// 0 仍允许并落默认 5/3，保持现有行为。
func TestCreateDatasourceFillsAvailabilityDefaultsWhenDisabled(t *testing.T) {
	sources := &fixDatasourceRepo{}
	gateway := fixDatasourceGateway(sources)
	if _, err := gateway.CreateDatasource(fixDatasourceContext(), fixAvailabilityDatasource(false, 0, 0)); err != nil {
		t.Fatalf("CreateDatasource error = %v", err)
	}
	if sources.created == nil {
		t.Fatal("repo was not called")
	}
	if sources.created.FailureWindowMinutes != defaultFailureWindowMinutes || sources.created.FailureThreshold != defaultFailureThreshold {
		t.Fatalf("defaults = %d/%d, want %d/%d", sources.created.FailureWindowMinutes, sources.created.FailureThreshold, defaultFailureWindowMinutes, defaultFailureThreshold)
	}
}

// TestCreateDatasourceAcceptsEnabledAvailability 校验合法启用请求按原值落库。
func TestCreateDatasourceAcceptsEnabledAvailability(t *testing.T) {
	sources := &fixDatasourceRepo{}
	gateway := fixDatasourceGateway(sources)
	if _, err := gateway.CreateDatasource(fixDatasourceContext(), fixAvailabilityDatasource(true, defaultFailureWindowMinutes, defaultFailureThreshold)); err != nil {
		t.Fatalf("CreateDatasource error = %v", err)
	}
	if sources.created == nil || sources.created.FailureWindowMinutes != defaultFailureWindowMinutes || sources.created.FailureThreshold != defaultFailureThreshold {
		t.Fatalf("created = %+v", sources.created)
	}
}

func fixExistingDatasource(enabled bool) *bo.Datasource {
	item := fixAvailabilityDatasource(enabled, defaultFailureWindowMinutes, defaultFailureThreshold)
	return &item
}

// TestUpdateDatasourceRejectsZeroAvailabilityWhenEnabled 覆盖 Update 路径的同一缺陷。
func TestUpdateDatasourceRejectsZeroAvailabilityWhenEnabled(t *testing.T) {
	cases := []struct {
		name      string
		window    int
		threshold int
	}{
		{name: "zero window", window: 0, threshold: 3},
		{name: "zero threshold", window: 5, threshold: 0},
		{name: "zero both", window: 0, threshold: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sources := &fixDatasourceRepo{existing: fixExistingDatasource(true)}
			gateway := fixDatasourceGateway(sources)
			if _, err := gateway.UpdateDatasource(fixDatasourceContext(), fixAvailabilityDatasource(true, tc.window, tc.threshold)); !merr.IsInvalid(err) {
				t.Fatalf("UpdateDatasource error = %v, want invalid", err)
			}
			if sources.updated != nil {
				t.Fatalf("repo should not be called, got %+v", sources.updated)
			}
		})
	}
}

// TestUpdateDatasourceFillsAvailabilityDefaultsWhenDisabled 校验 Update 路径关闭告警时落默认值。
func TestUpdateDatasourceFillsAvailabilityDefaultsWhenDisabled(t *testing.T) {
	sources := &fixDatasourceRepo{existing: fixExistingDatasource(false)}
	gateway := fixDatasourceGateway(sources)
	if _, err := gateway.UpdateDatasource(fixDatasourceContext(), fixAvailabilityDatasource(false, 0, 0)); err != nil {
		t.Fatalf("UpdateDatasource error = %v", err)
	}
	if sources.updated == nil {
		t.Fatal("repo was not called")
	}
	if sources.updated.FailureWindowMinutes != defaultFailureWindowMinutes || sources.updated.FailureThreshold != defaultFailureThreshold {
		t.Fatalf("defaults = %d/%d", sources.updated.FailureWindowMinutes, sources.updated.FailureThreshold)
	}
}

// TestUpdateDatasourceAcceptsEnabledAvailability 校验 Update 路径合法启用请求按原值落库。
func TestUpdateDatasourceAcceptsEnabledAvailability(t *testing.T) {
	sources := &fixDatasourceRepo{existing: fixExistingDatasource(true)}
	gateway := fixDatasourceGateway(sources)
	if _, err := gateway.UpdateDatasource(fixDatasourceContext(), fixAvailabilityDatasource(true, defaultFailureWindowMinutes, defaultFailureThreshold)); err != nil {
		t.Fatalf("UpdateDatasource error = %v", err)
	}
	if sources.updated == nil || sources.updated.FailureWindowMinutes != defaultFailureWindowMinutes || sources.updated.FailureThreshold != defaultFailureThreshold {
		t.Fatalf("updated = %+v", sources.updated)
	}
}
