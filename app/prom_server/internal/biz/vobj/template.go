package vobj

type NotifyTemplateType int32

const (
	// NotifyTemplateTypeCustom  自定义通知模板
	NotifyTemplateTypeCustom NotifyTemplateType = iota
	// NotifyTemplateTypeEmail 邮件通知模板
	NotifyTemplateTypeEmail
	// NotifyTemplateTypeSms 短信通知模板
	NotifyTemplateTypeSms
	// NotifyTemplateTypeWeChatWork 企业微信通知模板
	NotifyTemplateTypeWeChatWork
	// NotifyTemplateTypeFeiShu 飞书通知模板
	NotifyTemplateTypeFeiShu
	// NotifyTemplateTypeDingDing 钉钉通知模板
	NotifyTemplateTypeDingDing
)
