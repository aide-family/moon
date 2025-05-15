package do

import (
	"time"

	"gorm.io/plugin/soft_delete"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Resource interface {
	Base
	GetID() uint32
	GetName() string
	GetPath() string
	GetStatus() vobj.GlobalStatus
	GetAllow() vobj.ResourceAllow
	GetRemark() string
	GetMenus() []Menu
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() soft_delete.DeletedAt
}
