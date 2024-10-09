package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

type (

	// CreateAlarmNoticeGroupParams 创建告警组请求参数
	CreateAlarmNoticeGroupParams struct {
		// 告警组名称
		Name string `json:"name,omitempty"`
		// 告警组说明信息
		Remark string `json:"remark,omitempty"`
		// 告警组状态
		Status vobj.Status `json:"status,omitempty"`
		// 告警分组通知人
		NoticeMembers []*CreateNoticeMemberParams `json:"noticeMembers,omitempty"`
		// hook ids
		HookIds []uint32 `json:"hookIds"`
	}

	// CreateNoticeMemberParams 创建通知人参数
	CreateNoticeMemberParams struct {
		// 用户id
		MemberID uint32
		// 通知方式
		NotifyType vobj.NotifyType
	}

	// UpdateAlarmNoticeGroupStatusParams 更新告警组状态请求参数
	UpdateAlarmNoticeGroupStatusParams struct {
		IDs    []uint32 `json:"ids"`
		Status vobj.Status
	}

	// UpdateAlarmNoticeGroupParams 更新告警组请求参数
	UpdateAlarmNoticeGroupParams struct {
		ID          uint32 `json:"id"`
		UpdateParam *CreateAlarmNoticeGroupParams
	}

	// QueryAlarmNoticeGroupListParams 查询告警组列表请求参数
	QueryAlarmNoticeGroupListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Name    string
		Status  vobj.Status
	}

	// MyAlarmGroupListParams 我的告警组列表请求参数
	MyAlarmGroupListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Name    string
		Status  vobj.Status
	}

	// GetRealTimeAlarmParams 获取实时告警参数
	GetRealTimeAlarmParams struct {
		// 告警ID
		RealtimeAlarmID uint32
		// 告警指纹
		Fingerprint string
	}

	// GetRealTimeAlarmsParams 获取实时告警列表参数
	GetRealTimeAlarmsParams struct {
		// 分页参数
		Pagination types.Pagination
		// 告警时间范围
		EventAtStart int64
		EventAtEnd   int64
		// 告警恢复时间
		ResolvedAtStart int64
		ResolvedAtEnd   int64
		// 告警级别
		AlarmLevels []uint32
		// 告警状态
		AlarmStatuses []vobj.AlertStatus
		// 关键字
		Keyword string
		// 告警页面
		AlarmPageID uint32
		// 我的告警
		MyAlarm bool
	}

	// CreateAlarmItemParams 创建告警项请求参数
	CreateAlarmItemParams struct {
		// 告警状态, firing, resolved
		Status string `json:"status"`
		// 标签
		Labels map[string]string `json:"labels"`
		// 注解
		Annotations map[string]string `json:"annotations"`
		// 开始时间
		StartsAt int64 `json:"startsAt"`
		// 结束时间, 空表示未结束
		EndsAt int64 `json:"endsAt"`
		// 告警生成链接
		GeneratorURL string `json:"generatorURL"`
		// 指纹
		Fingerprint string `json:"fingerprint"`
		// 数据源ID
		DatasourceID uint32 `json:"datasourceID"`
	}

	// CreateAlarmInfoParams 创建告警信息参数
	CreateAlarmInfoParams struct {
		// 告警原始表id
		RawInfoID     uint32                          `json:"rawId"`
		TeamID        uint32                          `json:"teamId"`
		Alerts        []*CreateAlarmItemParams        `json:"alerts"`
		Strategy      *bizmodel.Strategy              `json:"strategy"`
		Level         *bizmodel.StrategyLevel         `json:"level"`
		DatasourceMap map[uint32]*bizmodel.Datasource `json:"datasourceMap"`
	}

	// CreateAlarmHookRawParams 告警hook原始信息
	CreateAlarmHookRawParams struct {
		Receiver          string                   `json:"receiver"`
		Status            string                   `json:"status"`
		GroupLabels       *vobj.Labels             `json:"groupLabels"`
		CommonLabels      *vobj.Labels             `json:"commonLabels"`
		CommonAnnotations map[string]string        `json:"commonAnnotations"`
		ExternalURL       string                   `json:"externalURL"`
		Version           string                   `json:"version"`
		GroupKey          string                   `json:"groupKey"`
		TruncatedAlerts   int32                    `json:"truncatedAlerts"`
		Fingerprint       string                   `json:"fingerprint"`
		Alerts            []*CreateAlarmItemParams `json:"alerts"`
		TeamID            uint32                   `json:"teamId"`
		StrategyID        uint32                   `json:"strategyId"`
		LevelID           uint32                   `json:"levelId"`
	}
)

func (a *CreateAlarmHookRawParams) String() string {
	if types.IsNil(a) {
		return ""
	}
	bs, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	return string(bs)
}

func (a *CreateAlarmHookRawParams) Index() string {
	return "palace:alert:hook:" + a.Fingerprint
}

func (a *CreateAlarmHookRawParams) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlert)
}

func (a *CreateAlarmInfoParams) GetDatasourceMap(datasourceID uint32) string {
	if types.IsNil(a) || types.IsNil(a.DatasourceMap) {
		return ""
	}
	if v, ok := a.DatasourceMap[datasourceID]; ok {
		return v.String()
	}
	return ""
}
