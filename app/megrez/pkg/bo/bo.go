package bo

import "time"

// RuleSnapshot 是领取到的一条规则。密码和密钥只在评测进程内使用。
type RuleSnapshot struct {
	UID           string
	NamespaceUID  string
	Name          string
	Type          string
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
type EvalReport struct {
	RuleUID  string
	EvalSlot time.Time
	Skipped  bool
	Hits     []EvalHit
}

// Sample 是一个时间序列。
type Sample struct {
	Labels map[string]string
	Points []Point
}

// Point 是一个采样点。
type Point struct {
	At    time.Time
	Value float64
	Raw   string
}
