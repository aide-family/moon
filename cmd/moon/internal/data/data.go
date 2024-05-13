package data

import (
	"context"

	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/moon/internal/conf"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	mainDB *gorm.DB
	bizDB  *gorm.DB
}

var closeFuncList []func()

// NewData .
func NewData(c *conf.Bootstrap) (*Data, func(), error) {
	logger := log.GetLogger()
	d := &Data{}

	mainConf := c.GetData().GetDatabase()
	bizConf := c.GetData().GetBizDatabase()

	if !types.IsNil(mainConf) && !types.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn.NewGormDB(mainConf.GetDsn(), mainConf.GetDriver())
		if err != nil {
			return nil, nil, err
		}
		d.mainDB = mainDB
		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			mainDBClose.Close()
		})
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) {
		bizDB, err := conn.NewGormDB(bizConf.GetDsn(), bizConf.GetDriver())
		if err != nil {
			return nil, nil, err
		}
		d.bizDB = bizDB
		closeFuncList = append(closeFuncList, func() {
			bizDBClose, _ := d.bizDB.DB()
			bizDBClose.Close()
		})
	}

	cleanup := func() {
		for _, f := range closeFuncList {
			f()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	return d, cleanup, nil
}

// GetMainDB 获取主库连接
func (d *Data) GetMainDB(ctx context.Context) *gorm.DB {
	db, exist := ctx.Value(conn.GormContextTxKey{}).(*gorm.DB)
	if exist {
		return db
	}
	return d.mainDB
}

// GetBizDB 获取业务库连接
func (d *Data) GetBizDB(ctx context.Context) *gorm.DB {
	db, exist := ctx.Value(conn.GormContextTxKey{}).(*gorm.DB)
	if exist {
		return db
	}
	return d.bizDB
}
