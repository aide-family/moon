package vobj

type ChangedType int32

const (
	ChangedTypeUnknown ChangedType = iota
	ChangedTypeMetricDatasource
	ChangedTypeMetricStrategy
	ChangedTypeNoticeGroup
	ChangedTypeNoticeSMSConfig
	ChangedTypeNoticeEmailConfig
	ChangedTypeNoticeHookConfig
)
