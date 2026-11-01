package bo

import "time"

// 探测驱动，取值与 pkg/probe、proto 的 driver 枚举一致。
const (
	ProbeDriverHTTP = "PROBE_HTTP"
	ProbeDriverTCP  = "PROBE_TCP"
	ProbeDriverICMP = "PROBE_ICMP"
)

// ProbeDrivers 按固定顺序返回全部探测驱动。调度与领取按「驱动属于探测类型」
// 查询，新增类型时只改这里，不在各处手写驱动清单。
func ProbeDrivers() []string {
	return []string{ProbeDriverHTTP, ProbeDriverTCP, ProbeDriverICMP}
}

// IsProbeDriver 判断驱动是否属于探测类型。
func IsProbeDriver(driver string) bool {
	switch driver {
	case ProbeDriverHTTP, ProbeDriverTCP, ProbeDriverICMP:
		return true
	default:
		return false
	}
}

// ProbeKindOf 把探测驱动映射为事件标签 probe_kind 的展示值。未知驱动返回空字符串。
func ProbeKindOf(driver string) string {
	switch driver {
	case ProbeDriverHTTP:
		return "HTTP"
	case ProbeDriverTCP:
		return "TCP"
	case ProbeDriverICMP:
		return "ICMP"
	default:
		return ""
	}
}

// 探测规则比较的观测指标。
const (
	ProbeMetricUp         = "UP"
	ProbeMetricLatencyMS  = "LATENCY_MS"
	ProbeMetricStatusCode = "STATUS_CODE"
	ProbeMetricCertDays   = "CERT_DAYS"
)

// ValidProbeMetric 判断观测指标是否适用于该驱动。HTTP 有状态码与证书，四个
// 指标都可用；TCP 与 ICMP 只探测连通性和耗时，只允许 UP 与 LATENCY_MS。
func ValidProbeMetric(driver, metric string) bool {
	switch driver {
	case ProbeDriverHTTP:
		switch metric {
		case ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricStatusCode, ProbeMetricCertDays:
			return true
		}
	case ProbeDriverTCP, ProbeDriverICMP:
		switch metric {
		case ProbeMetricUp, ProbeMetricLatencyMS:
			return true
		}
	}
	return false
}

// ProbeTargetGroup 是一组共用请求参数的探测地址。
type ProbeTargetGroup struct {
	UID          string
	NamespaceUID string
	Name         string
	// Driver 取 PROBE_HTTP、PROBE_TCP 或 PROBE_ICMP。
	Driver string
	Status string
	// Labels 是规则匹配目标组用的 JSON 对象。
	Labels string
	// ExecutorLabels 预留给边缘探测，空对象表示集群内执行。
	ExecutorLabels string
	// Request 是组内地址共用的请求参数 JSON，形态见 pkg/probe.Request。
	Request string
	Remark  string
	// Targets 是组内地址，按 Sort 升序。
	Targets []ProbeTarget
	// TargetCount 与 ActiveTargetCount 由仓储补齐，便于列表展示。
	TargetCount       int
	ActiveTargetCount int
	// LastProbedAt 是组内最近一次探测时间，还没有探测过时为空。
	LastProbedAt *time.Time
}

// ProbeTarget 是目标组里的一个地址。
type ProbeTarget struct {
	UID          string
	NamespaceUID string
	GroupUID     string
	// Spec 是归一化后的地址。
	Spec   string
	Labels string
	Status string
	Sort   int
	// LastResult 是该地址最近一次探测结果，只在详情里补齐。
	LastResult *ProbeResult
}

// ProbeResult 是一次探测的结果。
type ProbeResult struct {
	UID          string
	NamespaceUID string
	RuleUID      string
	RuleName     string
	GroupUID     string
	TargetUID    string
	TargetSpec   string
	// Slot 是这一轮的计划时刻，按探测周期对齐。
	Slot     time.Time
	ProbedAt time.Time
	OK       bool
	// Value 是与规则观测指标对应的数值文本。
	Value        string
	DurationMS   int
	StatusCode   int
	ErrorKind    string
	ErrorMessage string
	DNSMS        int
	TCPMS        int
	TLSMS        int
	TTFBMS       int
	CertDays     int
	ResolvedIP   string
	Executor     string
}

// ProbeTask 是一条待执行的探测：一个规则在一个目标组上、本轮要探测的目标。
type ProbeTask struct {
	Rule    AlertRule
	Group   ProbeTargetGroup
	Targets []ProbeTarget
	// Slot 是本轮计划时刻。
	Slot time.Time
}

// ProbeReportItem 是执行器回写的单个目标结果。
type ProbeReportItem struct {
	TargetUID    string
	OK           bool
	Value        string
	DurationMS   int
	StatusCode   int
	ErrorKind    string
	ErrorMessage string
	DNSMS        int
	TCPMS        int
	TLSMS        int
	TTFBMS       int
	CertDays     int
	ResolvedIP   string
	Executor     string
}

// ProbeReport 是执行器的一次回写。Slot 是领取时的计划时刻。
type ProbeReport struct {
	RuleUID string
	Slot    time.Time
	Results []ProbeReportItem
}
