package bo

import "time"

// RuleSnapshot 是领取到的一条规则。密码和密钥只在评测进程内使用。
type RuleSnapshot struct {
	UID           string
	NamespaceUID  string
	Name          string
	Driver        string
	Expr          string
	Labels        string
	Summary       string
	Description   string
	EvalSlot      time.Time
	DatasourceUID string
	Endpoint      string
	Username      string
	Password      string
	Secret        string
	// DatasourceName 只用于心跳里展示「规则 × 数据源」组合，不参与评测。
	DatasourceName string
	// IndexPattern 和 TimeField 只对 Elasticsearch 有意义：索引模式决定搜哪些索引，
	// 时间字段决定 date_histogram 按哪个字段切桶。空值分别表示全部索引和 @timestamp。
	IndexPattern string
	TimeField    string
	Levels       []Level
	Firing       []string
	// Variables 是摘要和描述模板里可用的自定义查询变量。
	Variables []Variable
}

// Variable 是一条自定义查询变量。Expr 是查询模板，展开时可用主查询命中
// 序列的标签（{{ $labels.xxx }}），再按数据源做即时查询。
type Variable struct {
	Name        string
	Expr        string
	Description string
}

// Level 是一条级别条件。
type Level struct {
	Condition  string
	Sample     string
	Threshold  string
	ForSeconds int
	LevelUID   string
}

// EvalHit 是一次命中。
type EvalHit struct {
	Fingerprint string
	Labels      string
	Value       string
	LevelUID    string
	Summary     string
	Description string
}

// 一个「规则 × 数据源」组合在本轮领取里的状态。
const (
	// WorkerTaskRunning 表示这个组合正在被评测。
	WorkerTaskRunning = "RUNNING"
	// WorkerTaskNew 表示本轮新加入、上一批没有这个组合。
	WorkerTaskNew = "NEW"
	// WorkerTaskPending 表示本轮已领取、还在排队等待评测。
	WorkerTaskPending = "PENDING"
	// WorkerTaskDone 表示本轮已经回写完成。
	WorkerTaskDone = "DONE"
	// WorkerTaskLeft 表示上一批领取过、本批已不再包含它。
	WorkerTaskLeft = "LEFT"
)

// WorkerTask 是一轮领取里的一个「规则 × 数据源」组合。
type WorkerTask struct {
	State          string
	RuleUID        string
	RuleName       string
	DatasourceUID  string
	DatasourceName string
	EvalSlot       time.Time
}

// WorkerStatus is the safe instance heartbeat sent to Alioth.
type WorkerStatus struct {
	InstanceUID         string
	Hostname            string
	Busy                bool
	TaskUID             string
	TaskName            string
	TaskStartedAt       time.Time
	Tasks               []WorkerTask
	RoundStartedAt      time.Time
	NextPollAt          time.Time
	PollIntervalSeconds int32
}

// EvalReport is a rule evaluation result.
// Failed marks a datasource request failure that alioth counts toward availability;
// Hits stays empty in that case and the failed batch never recovers rule events.
type EvalReport struct {
	RuleUID        string
	DatasourceUID  string
	EvalSlot       time.Time
	Skipped        bool
	Failed         bool
	FailureMessage string
	Hits           []EvalHit
}

// Sample 是一个时间序列。Step 是数据源的时间桶秒数，只有按桶聚合的结果
// （VictoriaLogs stats）才非 0；FOR 用它判断相邻采样点之间是否缺桶。
type Sample struct {
	Labels map[string]string
	Points []Point
	Step   int64
}

// Point 是一个采样点。
type Point struct {
	At    time.Time
	Value float64
	Raw   string
}
