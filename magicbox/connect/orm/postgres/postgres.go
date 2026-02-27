package postgres

import (
	"bytes"
	"fmt"
	"strings"

	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/connect/orm"
	"github.com/aide-family/magicbox/log/gormlog"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
)

func init() {
	connect.RegisterORMFactory(config.ORMConfig_POSTGRES, New)
}

// New creates a new PostgreSQL database connection.
// If the PostgreSQL options are not valid, it will return an error.
// The database connection is not closed, you need to call the returned function to close the connection.
func New(c *config.ORMConfig) (*gorm.DB, error) {
	postgresConf := &config.PostgreSQLOptions{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), postgresConf, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, merr.ErrorInternalServer("unmarshal postgres config failed: %v", err)
		}
	}
	ormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	if strings.EqualFold(c.GetUseSystemLogger(), "true") {
		ormConfig.Logger = gormlog.New(klog.GetLogger())
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", postgresConf.Host, postgresConf.Username, postgresConf.Password, postgresConf.Database, postgresConf.Port)
	dsnBuf := bytes.NewBufferString(dsn)
	for key, value := range postgresConf.Parameters {
		dsnBuf.WriteString(fmt.Sprintf(" %s=%s", key, value))
	}
	dsn = dsnBuf.String()
	b := &orm.Builder{
		Dialector: postgres.Open(dsn),
		Config:    ormConfig,
		IsDebug:   strings.EqualFold(c.GetDebug(), "true"),
	}
	return b.Build()
}
