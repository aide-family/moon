package bo

import "time"

// 探测类型。一期只有 HTTP 探测。
const (
	ProbeDriverHTTP = "PROBE_HTTP"
)

// 探测规则比较的观测指标。
const (
	ProbeMetricUp         = "UP"
	ProbeMetricLatencyMS  = "LATENCY_MS"
	ProbeMetricStatusCode = "STATUS_CODE"
	ProbeMetricCertDays   = "CERT_DAYS"
)

// ProbeTargetGroup 是一组共用请求参数的探测地址。
type ProbeTargetGroup struct {
	UID          string
	NamespaceUID string
	Name         string
	// Driver 取 PROBE_HTTP。
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
