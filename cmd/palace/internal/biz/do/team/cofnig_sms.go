package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/crypto"
)

var _ do.TeamSMSConfig = (*SmsConfig)(nil)

const tableNameConfigSMS = "team_config_sms"

type SmsConfig struct {
	do.TeamModel
	Name     string                  `gorm:"column:name;type:varchar(32);not null;comment:名称" json:"name"`
	Remark   string                  `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	Status   vobj.GlobalStatus       `gorm:"column:status;type:tinyint;not null;default:0;comment:状态" json:"status"`
	Provider vobj.SMSProviderType    `gorm:"column:provider;type:tinyint(2);not null;comment:短信服务商" json:"provider"`
	Sms      *crypto.Object[*do.SMS] `gorm:"column:sms;type:text;not null;comment:短信配置" json:"sms"`
}

func (s *SmsConfig) GetProviderType() vobj.SMSProviderType {
	if s == nil {
		return vobj.SMSProviderTypeUnknown
	}
	return s.Provider
}

func (s *SmsConfig) GetName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *SmsConfig) GetRemark() string {
	if s == nil {
		return ""
	}
	return s.Remark
}

func (s *SmsConfig) GetStatus() vobj.GlobalStatus {
	if s == nil {
		return vobj.GlobalStatusUnknown
	}
	return s.Status
}

func (s *SmsConfig) GetSMSConfig() *do.SMS {
	if s == nil || s.Sms == nil {
		return nil
	}
	return s.Sms.Get()
}

func (s *SmsConfig) TableName() string {
	return tableNameConfigSMS
}
