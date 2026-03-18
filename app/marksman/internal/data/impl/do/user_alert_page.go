package do

import (
	"github.com/bwmarrin/snowflake"
)

// UserAlertPage stores a user's followed alert page (one row per user-page pair; order by sort_order).
type UserAlertPage struct {
	EventBaseModel
	NamespaceUID snowflake.ID `gorm:"column:namespace_uid;index"`
	UserUID      snowflake.ID `gorm:"column:user_uid;index:idx_user_uid"`
	AlertPageUID snowflake.ID `gorm:"column:alert_page_uid;index"`
	SortOrder    int32        `gorm:"column:sort_order;default:0"`
}

func (UserAlertPage) TableName() string {
	return "user_alert_pages"
}
