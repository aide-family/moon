package do

import (
	"errors"
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Namespace struct {
	UID       snowflake.ID                `gorm:"column:uid;not null;primaryKey"`
	CreatedAt time.Time                   `gorm:"column:created_at;type:datetime;not null;"`
	UpdatedAt time.Time                   `gorm:"column:updated_at;type:datetime;not null;"`
	DeletedAt gorm.DeletedAt              `gorm:"column:deleted_at;type:datetime;uniqueIndex:idx__namespace__name__deleted_at"`
	Creator   snowflake.ID                `gorm:"column:creator;not null;index"`
	Name      string                      `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx__namespace__name__deleted_at;default:''"`
	Metadata  *safety.Map[string, string] `gorm:"column:metadata;type:json;"`
	Status    enum.GlobalStatus           `gorm:"column:status;type:tinyint;not null;default:0"`
	Logo      string                      `gorm:"column:logo;type:varchar(100);not null;default:''"`
	Secret    string                      `gorm:"column:secret;type:varchar(100);not null;default:''"`
	Banners   string                      `gorm:"column:banners;type:varchar(500);not null;default:'';comment:'banner images, separated by comma'"`
	Remark    string                      `gorm:"column:remark;type:varchar(200);not null;default:''"`
	Leader    snowflake.ID                `gorm:"column:leader;not null;index;default:0"`
}

func (Namespace) TableName() string {
	return "namespaces"
}

func (n *Namespace) WithCreator(creator snowflake.ID) *Namespace {
	n.Creator = creator
	return n
}

func (n *Namespace) BeforeCreate(tx *gorm.DB) (err error) {
	if n.Creator == 0 {
		return errors.New("creator is required")
	}

	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return err
	}
	n.UID = node.Generate()
	if n.Status <= 0 {
		return errors.New("status is required")
	}
	return
}

func (n *Namespace) GetBanners() []string {
	if n.Banners == "" {
		return nil
	}
	return strings.Split(n.Banners, ",")
}
