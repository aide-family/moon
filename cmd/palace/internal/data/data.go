package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/data/query/systemgen"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/gorm"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/validate"
)

// ProviderSetData is a set of data providers.
var ProviderSetData = wire.NewSet(New)

// New a data and returns.
func New(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	var err error
	data := &Data{
		dataConf:   c.GetData(),
		mainDB:     nil,
		bizDBMap:   safety.NewMap[uint32, gorm.DB](),
		eventDBMap: safety.NewMap[uint32, gorm.DB](),
		cache:      nil,
		rabbitConn: safety.NewMap[string, *bo.Server](),
		houyiConn:  safety.NewMap[string, *bo.Server](),
		laurelConn: safety.NewMap[string, *bo.Server](),
		helper:     log.NewHelper(log.With(logger, "module", "data")),
	}

	dataConf := c.GetData()

	data.mainDB, err = gorm.NewDB(dataConf.GetMain())
	if err != nil {
		return nil, nil, err
	}
	data.cache, err = cache.NewCache(c.GetCache())
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		var key cache.K = "palace:table:exists"
		do.RegisterHasTableFunc(func(teamId uint32, tableName string) bool {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return data.cache.Client().HExists(ctx, key.Key(), tableName).Val()
		})
		do.RegisterCacheTableFlag(func(teamId uint32, tableName string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return data.cache.Client().HSet(ctx, key.Key(), tableName, true).Err()
		})
	}()

	return data, func() {
		data.helper.Info("closing the data resources")
		if err = data.mainDB.Close(); err != nil {
			data.helper.Errorw("method", "close main db", "err", err)
		}
		if err = data.cache.Close(); err != nil {
			data.helper.Errorw("method", "close cache", "err", err)
		}
		for teamID, db := range data.bizDBMap.List() {
			if err = db.Close(); err != nil {
				method := fmt.Sprintf("close team [%d] biz db", teamID)
				data.helper.Errorw("method", method, "err", err)
			}
		}
		for teamID, db := range data.eventDBMap.List() {
			if err = db.Close(); err != nil {
				method := fmt.Sprintf("close team [%d] alarm db", teamID)
				data.helper.Errorw("method", method, "err", err)
			}
		}
		for _, server := range data.rabbitConn.List() {
			if err = server.Conn.Close(); err != nil {
				data.helper.Errorw("method", "close rabbit conn", "err", err)
			}
		}
		for _, server := range data.houyiConn.List() {
			if err = server.Conn.Close(); err != nil {
				data.helper.Errorw("method", "close houyi conn", "err", err)
			}
		}
		for _, server := range data.laurelConn.List() {
			if err = server.Conn.Close(); err != nil {
				data.helper.Errorw("method", "close laurel conn", "err", err)
			}
		}
		if err = safety.Wait(); err != nil {
			data.helper.Errorw("method", "safety.Wait", "err", err)
		}
	}, nil
}

type Data struct {
	dataConf             *conf.Data
	mainDB               gorm.DB
	bizDBMap, eventDBMap *safety.Map[uint32, gorm.DB]
	cache                cache.Cache
	rabbitConn           *safety.Map[string, *bo.Server]
	houyiConn            *safety.Map[string, *bo.Server]
	laurelConn           *safety.Map[string, *bo.Server]
	helper               *log.Helper
}

func (d *Data) GetServerConn(serverType vobj.ServerType, id string) (*bo.Server, bool) {
	switch serverType {
	case vobj.ServerTypeRabbit:
		return d.rabbitConn.Get(id)
	case vobj.ServerTypeHouyi:
		return d.houyiConn.Get(id)
	case vobj.ServerTypeLaurel:
		return d.laurelConn.Get(id)
	}
	return nil, false
}
func (d *Data) SetServerConn(serverType vobj.ServerType, id string, conn *bo.Server) {
	switch serverType {
	case vobj.ServerTypeRabbit:
		d.rabbitConn.Set(id, conn)
	case vobj.ServerTypeHouyi:
		d.houyiConn.Set(id, conn)
	case vobj.ServerTypeLaurel:
		d.laurelConn.Set(id, conn)
	}
}

func (d *Data) RemoveServerConn(serverType vobj.ServerType, id string) {
	switch serverType {
	case vobj.ServerTypeRabbit:
		d.rabbitConn.Delete(id)
	case vobj.ServerTypeHouyi:
		d.houyiConn.Delete(id)
	case vobj.ServerTypeLaurel:
		d.laurelConn.Delete(id)
	}
}

func (d *Data) FirstServerConn(serverType vobj.ServerType) (*bo.Server, bool) {
	switch serverType {
	case vobj.ServerTypeRabbit:
		return d.rabbitConn.First()
	case vobj.ServerTypeHouyi:
		return d.houyiConn.First()
	case vobj.ServerTypeLaurel:
		return d.laurelConn.First()
	}
	return nil, false
}

func (d *Data) GetCache() cache.Cache {
	return d.cache
}

func (d *Data) GetMainDB() gorm.DB {
	return d.mainDB
}

func (d *Data) GetBizDB(teamID uint32) (gorm.DB, error) {
	db, ok := d.bizDBMap.Get(teamID)
	if ok {
		return db, nil
	}
	teamDo, err := d.queryTeam(teamID)
	if err != nil {
		return nil, err
	}
	dbConfig := teamDo.GetBizDBConfig()
	if validate.IsNil(dbConfig) {
		return d.GetMainDB(), nil
	}

	db, err = gorm.NewDB(dbConfig)
	if err != nil {
		return nil, merr.ErrorInternalServer("new team biz db err").WithCause(err)
	}
	d.bizDBMap.Set(teamID, db)
	return db, nil
}

func (d *Data) GetEventDB(teamID uint32) (gorm.DB, error) {
	db, ok := d.eventDBMap.Get(teamID)
	if ok {
		return db, nil
	}
	teamDo, err := d.queryTeam(teamID)
	if err != nil {
		return nil, err
	}
	dbConfig := teamDo.GetAlarmDBConfig()
	if validate.IsNil(dbConfig) {
		return d.GetMainDB(), nil
	}
	db, err = gorm.NewDB(dbConfig)
	if err != nil {
		return nil, merr.ErrorInternalServer("new team alarm db err").WithCause(err)
	}
	d.eventDBMap.Set(teamID, db)
	return db, nil
}

func (d *Data) queryTeam(teamID uint32) (*system.Team, error) {
	teamQuery := systemgen.Use(d.GetMainDB().GetDB()).Team
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	teamDo, err := teamQuery.WithContext(ctx).Where(teamQuery.ID.Eq(teamID)).First()
	if err != nil {
		return nil, merr.ErrorInternalServer("team query err").WithCause(err)
	}
	return teamDo, nil
}
