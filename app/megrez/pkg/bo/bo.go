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
	Levels        []Level
	Firing        []string
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

// EvalReport 是回写内容。Skipped 表示驱动尚未实现。
// WorkerStatus is the safe instance heartbeat sent to Alioth.
type WorkerStatus struct {
	InstanceUID   string
	Hostname      string
	Busy          bool
	TaskUID       string
	TaskName      string
	TaskStartedAt time.Time
}

// EvalReport is a rule evaluation result.
type EvalReport struct {
	RuleUID       string
	DatasourceUID string
	EvalSlot      time.Time
	Skipped       bool
	Hits          []EvalHit
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
