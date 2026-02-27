// Package sqlite provides a SQLite ORM factory.
package sqlite

import (
	"strings"

	"github.com/glebarez/sqlite"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/connect/orm"
	"github.com/aide-family/magicbox/log/gormlog"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
)

func init() {
	connect.RegisterORMFactory(config.ORMConfig_SQLITE, New)
}

// New creates a new SQLite database connection.
// If the SQLite options are not valid, it will return an error.
// The database connection is not closed, you need to call the returned function to close the connection.
func New(c *config.ORMConfig) (*gorm.DB, error) {
	sqliteConf := &config.SQLiteOptions{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), sqliteConf, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, merr.ErrorInternalServer("unmarshal sqlite config failed: %v", err)
		}
	}
	ormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	if strings.EqualFold(c.GetUseSystemLogger(), "true") {
		ormConfig.Logger = gormlog.New(klog.GetLogger())
	}
	b := &orm.Builder{
		Dialector: sqlite.Open(sqliteConf.Dsn),
		Config:    ormConfig,
		IsDebug:   strings.EqualFold(c.GetDebug(), "true"),
	}
	return b.Build()
}
