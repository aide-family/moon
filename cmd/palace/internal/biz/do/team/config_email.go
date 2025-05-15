package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/crypto"
)

var _ do.TeamEmailConfig = (*EmailConfig)(nil)

const tableNameConfigEmail = "team_config_emails"

// EmailConfig represents email configuration for a team
type EmailConfig struct {
	do.TeamModel
	Name   string                    `gorm:"column:name;type:varchar(20);not null;comment:配置名称" json:"name"`
	Remark string                    `gorm:"column:remark;type:varchar(200);not null;comment:配置备注" json:"remark"`
	Status vobj.GlobalStatus         `gorm:"column:status;type:tinyint(2);not null;default:0;comment:状态" json:"status"`
	Email  *crypto.Object[*do.Email] `gorm:"column:email;type:text;not null;comment:邮件配置" json:"email"`
}

func (c *EmailConfig) GetEmailConfig() *do.Email {
	if c == nil || c.Email == nil {
		return nil
	}
	return c.Email.Get()
}

func (c *EmailConfig) GetUser() string {
	if c == nil {
		return ""
	}
	emailConfig := c.GetEmailConfig()
	if emailConfig == nil {
		return ""
	}
	return emailConfig.User
}

func (c *EmailConfig) GetPass() string {
	if c == nil {
		return ""
	}
	emailConfig := c.GetEmailConfig()
	if emailConfig == nil {
		return ""
	}
	return emailConfig.Pass
}

func (c *EmailConfig) GetHost() string {
	if c == nil {
		return ""
	}
	emailConfig := c.GetEmailConfig()
	if emailConfig == nil {
		return ""
	}
	return emailConfig.Host
}

func (c *EmailConfig) GetPort() uint32 {
	if c == nil {
		return 0
	}
	emailConfig := c.GetEmailConfig()
	if emailConfig == nil {
		return 0
	}
	return emailConfig.Port
}

func (c *EmailConfig) GetEnable() bool {
	if c == nil {
		return false
	}
	return c.Status.IsEnable()
}

func (c *EmailConfig) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

func (c *EmailConfig) GetRemark() string {
	if c == nil {
		return ""
	}
	return c.Remark
}

func (c *EmailConfig) GetStatus() vobj.GlobalStatus {
	if c == nil {
		return vobj.GlobalStatusUnknown
	}
	return c.Status
}

func (c *EmailConfig) TableName() string {
	return tableNameConfigEmail
}
