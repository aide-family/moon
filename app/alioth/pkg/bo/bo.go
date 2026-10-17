package bo

import "time"

// Account 是全局用户。PasswordHash 只在登录校验时使用，离开用例前要清空。
type Account struct {
	UID          string
	Name         string
	Nickname     string
	Email        string
	Avatar       string
	Remark       string
	Status       string
	PasswordHash string
}

// Namespace 是租户。OwnerUID 指向当前负责人。
type Namespace struct {
	UID      string
	Name     string
	OwnerUID string
	Logo     string
	Remark   string
	Status   string
}

// Member 是用户在某个租户中的成员身份。
type Member struct {
	UID          string
	NamespaceUID string
	UserUID      string
	Email        string
	Name         string
	Status       string
	RoleUIDs     []string
	RoleNames    []string
}

// Role 是租户内角色。IsSystem 为真时不能改名、删除或改授权。
type Role struct {
	UID          string
	NamespaceUID string
	Name         string
	Status       string
	IsSystem     bool
	Permissions  []string
}

// Permission 是全局权限定义。
type Permission struct {
	UID      string
	Name     string
	Resource string
	Action   string
}

// SelectItem 是下拉列表中的一项。
type SelectItem struct {
	UID  string
	Name string
}

// Session 是登录成功后交给调用方的令牌对。User 不含密码哈希。
type Session struct {
	AccessToken  string
	RefreshToken string
	User         Account
}

// Datasource 是租户里的一条数据源。Password 和 Secret 只在用例内部是明文，离开接口前要清空。
type Datasource struct {
	UID          string
	NamespaceUID string
	Name         string
	Type         string
	Driver       string
	Endpoint     string
	Username     string
	Password     string
	Secret       string
	SecretSet    bool
	Status       string
	LevelUID     string
	LevelName    string
	Labels       string
	Remark       string
}

// AlertDictionary 是告警字典中的一条。类型为告警级别或数据源级别。
type AlertDictionary struct {
	UID          string
	NamespaceUID string
	Name         string
	Color        string
	Type         string
	Status       string
}

// AlertRuleGroup 是租户内的规则组。名称在未删除记录上唯一。
type AlertRuleGroup struct {
	UID          string
	NamespaceUID string
	Name         string
	Status       string
}

// AlertRuleLevel 是规则上的一条比较条件。
type AlertRuleLevel struct {
	UID        string
	Condition  string
	Sample     string
	Threshold  string
	ForSeconds int
	LevelUID   string
	LevelName  string
}

// AlertRule 是规则组里的一条规则。停用规则不会进入到期候选。
type AlertRule struct {
	UID                   string
	NamespaceUID          string
	GroupUID              string
	Name                  string
	Type                  string
	Driver                string
	Status                string
	Expr                  string
	Labels                string
	DatasourceUID         string
	DatasourceLabels      string
	Summary               string
	Description           string
	NextEvalAt            *time.Time
	ScheduledSlot         *time.Time
	Levels                []AlertRuleLevel
	NotificationGroupUIDs []string
}

// Due 判断规则此刻是否该被调度。停用规则不进入候选。日志和链路规则可以进入候选，由评测方标明未评测。
func (r AlertRule) Due(now time.Time) bool {
	if r.Status != "ACTIVE" {
		return false
	}
	return r.NextEvalAt == nil || !r.NextEvalAt.After(now)
}

// AlertEvent 是一条实时或历史告警。
type AlertEvent struct {
	UID           string
	NamespaceUID  string
	RuleUID       string
	DatasourceUID string
	Fingerprint   string
	Status        string
	Expr          string
	Labels        string
	Summary       string
	Description   string
	SampleValue   string
	LevelUID      string
	StartsAt      time.Time
	EndsAt        *time.Time
	EvalSlot      time.Time
	Suppressed    bool
	IntervenedBy  string
	CreatedAt     time.Time
}

// AlertStats 是当前租户的告警数量。
type AlertStats struct {
	Firing     int
	Suppressed int
	Recovered  int
}

// AlertPage 是实时告警的标签过滤视图。
type AlertPage struct {
	UID          string
	NamespaceUID string
	Name         string
	OwnerUID     string
	Labels       string
	Public       bool
	Remark       string
}

// RealtimeAlertTab 是实时告警页上的一个 Tab。
type RealtimeAlertTab struct {
	ID   string
	Name string
	Kind string
}

// AlertEventListFilter 是告警事件列表的过滤条件。
type AlertEventListFilter struct {
	View          string
	Name          string
	RuleUIDs      []string
	FilterByRules bool
	LabelsJSON    string
}

// EvalHit 是一次评测命中的序列。
type EvalHit struct {
	Fingerprint string
	Labels      string
	Value       string
	LevelUID    string
	Summary     string
	Description string
}

// EvalReport 是 megrez 回写的一次评测结果。Skipped 表示驱动尚未实现。
type EvalReport struct {
	RuleUID       string
	DatasourceUID string
	EvalSlot      time.Time
	Skipped       bool
	Hits          []EvalHit
}

// RuleSnapshot 是领取到的规则，含解密后的数据源连接信息。
type RuleSnapshot struct {
	Rule               AlertRule
	Datasource         Datasource
	FiringFingerprints []string
}

// NotificationChannel 是一条通知渠道。Password 和 Secret 只在写入时携带明文。
type NotificationChannel struct {
	UID          string
	NamespaceUID string
	Name         string
	Type         string
	App          string
	Status       string
	Endpoint     string
	Username     string
	Password     string
	Secret       string
	FromAddress  string
	Remark       string
	HasPassword  bool
	HasSecret    bool
}

// NotificationTemplate 是一条通知模板。
type NotificationTemplate struct {
	UID          string
	NamespaceUID string
	Name         string
	Type         string
	Status       string
	Subject      string
	Body         string
	Remark       string
}

// NotificationGroup 是一条通知组。
type NotificationGroup struct {
	UID          string
	NamespaceUID string
	Name         string
	Status       string
	ChannelUID   string
	TemplateUID  string
	Receivers    string
	Remark       string
}

// NotificationTask 是一条投递任务。
type NotificationTask struct {
	UID          string
	NamespaceUID string
	EventUID     string
	Phase        string
	Status       string
	RuleUID      string
	GroupUID     string
	ChannelUID   string
	TemplateUID  string
	Receivers    string
	Subject      string
	Body         string
	Params       string
	RetriesLeft  int
	LastError    string
}

// MessageLog 是一次投递结果。
type MessageLog struct {
	UID          string
	NamespaceUID string
	TaskUID      string
	Status       string
	ChannelUID   string
	TemplateUID  string
	Receivers    string
	Detail       string
	CreatedAt    time.Time
}

// NotifyReport 是 merak 回写的投递结果。
type NotifyReport struct {
	TaskUID string
	OK      bool
	Detail  string
}

// Agent 是一条边缘机器。Token 只在签发时返回明文。
type Agent struct {
	UID           string
	NamespaceUID  string
	Name          string
	Hostname      string
	MachineID     string
	LocalIP       string
	Version       string
	CollectJSON   string
	Online        bool
	Status        string
	SSHHost       string
	SSHPort       int
	SSHUser       string
	SSHPassword   string
	SSHPrivateKey string
	HasPassword   bool
	HasPrivateKey bool
	Remark        string
	Token         string
}

// SSHCommand 是一条 SSH 命令。
type SSHCommand struct {
	UID          string
	NamespaceUID string
	Name         string
	Command      string
	Workdir      string
	Env          string
	ReviewStatus string
	Remark       string
}

// SSHCommandAudit 是一条 SSH 审计。
type SSHCommandAudit struct {
	UID          string
	NamespaceUID string
	CommandUID   string
	AgentUID     string
	OperatorUID  string
	Kind         string
	Status       string
	Command      string
	ExitCode     int
	Output       string
	Truncated    bool
	Detail       string
	CreatedAt    time.Time
}

// AgentExecResult 是 Agent 回传的执行结果。
type AgentExecResult struct {
	RequestUID string
	ExitCode   int
	Output     string
	Truncated  bool
	Error      string
}

// SFTPAudit 是一条 SFTP 审计。
type SFTPAudit struct {
	UID          string
	NamespaceUID string
	AgentUID     string
	OperatorUID  string
	Path         string
	Action       string
	Result       string
	Detail       string
	CreatedAt    time.Time
}

// SFTPEntry 是一条目录项。
type SFTPEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	Mode    string
	ModTime string
}

// K8sProxyAudit 是一条 Kubernetes 代理审计。
type K8sProxyAudit struct {
	UID           string
	NamespaceUID  string
	AgentUID      string
	OperatorUID   string
	Method        string
	Resource      string
	KubeNamespace string
	Result        string
	Detail        string
	CreatedAt     time.Time
}

// NotifySnapshot 是 merak 领取到的任务，含解密后的渠道密钥。
type NotifySnapshot struct {
	UID             string
	NamespaceUID    string
	Phase           string
	ChannelUID      string
	ChannelType     string
	ChannelApp      string
	ChannelEndpoint string
	ChannelUsername string
	ChannelPassword string
	ChannelSecret   string
	ChannelFrom     string
	TemplateUID     string
	TemplateType    string
	TemplateSubject string
	TemplateBody    string
	Receivers       string
	Params          string
	Subject         string
	Body            string
}
