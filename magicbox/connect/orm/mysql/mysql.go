// Package mysql provides a MySQL ORM factory.
package mysql

import (
	"fmt"
	"net/url"
	"strings"

	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/connect/orm"
	"github.com/aide-family/magicbox/log/gormlog"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
)

func init() {
	connect.RegisterORMFactory(config.ORMConfig_MYSQL, New)
}

// New creates a new MySQL database connection.
// If the MySQL options are not valid, it will return an error.
// The database connection is not closed, you need to call the returned function to close the connection.
func New(c *config.ORMConfig) (*gorm.DB, error) {
	mysqlConf := &config.MySQLOptions{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), mysqlConf, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, merr.ErrorInternalServer("unmarshal mysql config failed: %v", err)
		}
	}
	params := url.Values{}
	for key, value := range mysqlConf.Parameters {
		params.Add(key, value)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", mysqlConf.Username, mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.Database, params.Encode())
	ormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	if strings.EqualFold(c.GetUseSystemLogger(), "true") {
		ormConfig.Logger = gormlog.New(klog.GetLogger())
	}

	b := &orm.Builder{
		Dialector: mysql.Open(dsn),
		Config:    ormConfig,
		IsDebug:   strings.EqualFold(c.GetDebug(), "true"),
	}
	return b.Build()
}
