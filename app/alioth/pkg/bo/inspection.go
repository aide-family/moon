package bo

import "time"

// InspectionPlan 是一次周期巡检的配置，包含检查项、报告模板、调度与通知组。
type InspectionPlan struct {
	UID                   string
	NamespaceUID          string
	Name                  string
	Description           string
	Status                string
	Cron                  string
	Timezone              string
	TimeoutSeconds        int
	TemplateUID           string
	OutputFormats         []string
	TitleTemplate         string
	NotificationGroupUIDs []string
	NotifyPolicy          string
	Labels                string
	Remark                string
	NextRunAt             *time.Time
	LastRunAt             *time.Time
	CreatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Items                 []InspectionItem
}

// InspectionItem 是计划里的一个检查项。
type InspectionItem struct {
	UID           string
	NamespaceUID  string
	PlanUID       string
	Kind          string
	Title         string
	Sort          int
	DatasourceUID string
	Query         string
	WindowSeconds int
	StepSeconds   int
	Display       string
	Params        string
	Enabled       bool
}

// InspectionPlanListFilter 是巡检计划列表的查询条件。
type InspectionPlanListFilter struct {
	Name        string
	Status      string
	TemplateUID string
}

// InspectionTemplate 是报告版式：一份模板两段正文。
type InspectionTemplate struct {
	UID          string
	NamespaceUID string
	Name         string
	Description  string
	HTMLBody     string
	MarkdownBody string
	IsBuiltin    bool
	Status       string
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InspectionRun 是一次巡检执行记录。
type InspectionRun struct {
	UID          string
	NamespaceUID string
	PlanUID      string
	PlanName     string
	Trigger      string
	Status       string
	ScheduledAt  time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
	WindowStart  *time.Time
	WindowEnd    *time.Time
	ItemTotal    int
	ItemOK       int
	ItemWarn     int
	ItemFailed   int
	Summary      string
	Error        string
	OperatorUID  string
	OperatorName string
	NotifiedAt   *time.Time
	CreatedAt    time.Time
}

// InspectionRunListFilter 是执行记录的查询条件。
type InspectionRunListFilter struct {
	PlanUID string
	Status  string
	Trigger string
	From    *time.Time
	To      *time.Time
}

// InspectionReport 是一次执行渲染出的报告快照。
type InspectionReport struct {
	UID             string
	NamespaceUID    string
	RunUID          string
	Format          string
	Title           string
	Content         string
	Size            int
	Checksum        string
	ShareEnabled    bool
	ShareExpiresAt  *time.Time
	ShareCreatedBy  string
	ShareCreatedAt  *time.Time
	ShareViewCount  int
	ShareLastViewed *time.Time
	CreatedAt       time.Time
}
