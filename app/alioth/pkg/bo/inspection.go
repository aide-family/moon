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
	Format       string
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
	Result       string
	Error        string
	OperatorUID  string
	OperatorName string
	NotifiedAt   *time.Time
	LeaseOwner   string
	LeaseUntil   *time.Time
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
	UID               string
	NamespaceUID      string
	RunUID            string
	Format            string
	Title             string
	Content           string
	Size              int
	Checksum          string
	ShareEnabled      bool
	ShareTokenHash    string
	SharePasswordHash string
	ShareExpiresAt    *time.Time
	ShareCreatedBy    string
	ShareCreatedAt    *time.Time
	ShareViewCount    int
	ShareLastViewedAt *time.Time
	CreatedAt         time.Time
}

// InspectionSection 是一个检查项的执行结果，模板只消费这个结构。
type InspectionSection struct {
	Title   string                `json:"title"`
	Kind    string                `json:"kind"`
	Status  string                `json:"status"`
	Error   string                `json:"error,omitempty"`
	Note    string                `json:"note,omitempty"`
	Metrics []InspectionMetric    `json:"metrics,omitempty"`
	Table   *InspectionTable      `json:"table,omitempty"`
	Series  []InspectionSeries    `json:"series,omitempty"`
	Logs    []InspectionLogSample `json:"logs,omitempty"`
	Alerts  []InspectionAlertRow  `json:"alerts,omitempty"`
	Raw     string                `json:"raw,omitempty"`
}

// InspectionMetric 是报告中的指标卡。
type InspectionMetric struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

// InspectionTable 是报告中的表格。
type InspectionTable struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

// InspectionSeries 是一条趋势序列。
type InspectionSeries struct {
	Name   string            `json:"name"`
	Points []InspectionPoint `json:"points"`
}

// InspectionPoint 是趋势序列上的一个点。
type InspectionPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// InspectionLogSample 是日志检查项的一条样本。
type InspectionLogSample struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// InspectionAlertRow 是告警检查项的一行事件。
type InspectionAlertRow struct {
	Title     string `json:"title"`
	Level     string `json:"level"`
	Status    string `json:"status"`
	StartedAt string `json:"startedAt"`
}

// InspectionResult 是执行记录的 result JSON。
type InspectionResult struct {
	Sections []InspectionSection `json:"sections"`
}

// InspectionSummary 是执行记录的摘要 JSON。
type InspectionSummary struct {
	Total    int `json:"total"`
	OK       int `json:"ok"`
	Warn     int `json:"warn"`
	Failed   int `json:"failed"`
	Abnormal int `json:"abnormal"`
	Duration int `json:"durationMs"`
}

// InspectionReportData 是报告模板的数据契约，HTML 与 Markdown 共用同一份。
type InspectionReportData struct {
	Title       string
	GeneratedAt time.Time
	Tenant      ReportTenant
	Plan        ReportPlan
	Run         ReportRun
	Summary     InspectionSummary
	Sections    []InspectionSection
	Abnormal    []InspectionSection
	HasAbnormal bool
}

// ReportTenant 是报告里的租户信息。
type ReportTenant struct {
	Name string
}

// ReportPlan 是报告里的计划信息。
type ReportPlan struct {
	Name        string
	Description string
	Labels      string
}

// ReportRun 是报告里的执行信息。
type ReportRun struct {
	Trigger      string
	StartedAt    time.Time
	FinishedAt   time.Time
	WindowStart  time.Time
	WindowEnd    time.Time
	Duration     string
	OperatorName string
}

// InspectionReportListFilter 是报告列表的查询条件。
type InspectionReportListFilter struct {
	RunUID string
	Format string
	From   *time.Time
	To     *time.Time
}
