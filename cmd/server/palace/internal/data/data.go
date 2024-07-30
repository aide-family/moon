package data

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/conn/cacher/nutsdbcacher"
	"github.com/aide-family/moon/pkg/util/conn/cacher/rediscacher"
	"github.com/aide-family/moon/pkg/util/conn/rbac"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm/clause"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	bizDatabaseConf   *palaceconf.Data_Database
	alarmDatabaseConf *palaceconf.Data_Database

	mainDB       *gorm.DB
	alarmDB      *sql.DB
	bizDB        *sql.DB
	cacher       conn.Cache
	enforcerMap  *sync.Map
	teamBizDBMap *sync.Map
	alarmDBMap   *sync.Map

	exit chan struct{}
}

var closeFuncList []func()

// NewData .
func NewData(c *palaceconf.Bootstrap) (*Data, func(), error) {
	mainConf := c.GetData().GetDatabase()
	alarmConf := c.GetData().GetAlarmDatabase()
	bizConf := c.GetData().GetBizDatabase()
	cacheConf := c.GetData().GetCache()
	d := &Data{
		bizDatabaseConf:   bizConf,
		alarmDatabaseConf: alarmConf,
		teamBizDBMap:      new(sync.Map),
		alarmDBMap:        new(sync.Map),
		enforcerMap:       new(sync.Map),
		exit:              make(chan struct{}),
	}
	cleanup := func() {
		for _, f := range closeFuncList {
			f()
		}
		log.Info("closing the data resources")
	}
	closeFuncList = append(closeFuncList, func() {
		log.Debug("close data")
		d.exit <- struct{}{}
	})

	if !types.IsNil(cacheConf) {
		d.cacher = newCache(cacheConf)
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close cache", d.cacher.Close())
		})
	}

	if !types.IsNil(alarmConf) && !types.TextIsNull(alarmConf.GetDsn()) {
		// 打开数据库连接
		db, err := sql.Open(alarmConf.GetDriver(), bizConf.GetDsn())
		if !types.IsNil(err) {
			log.Fatalf("Error opening database: %v\n", err)
			cleanup()
			return nil, nil, err
		}
		d.alarmDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close alarm db", d.bizDB.Close())
			d.alarmDBMap.Range(func(key, value any) bool {
				alarmDB, ok := value.(*gorm.DB)
				if !ok {
					return true
				}
				dbTmp, _ := alarmDB.DB()
				log.Debugw("close alarm orm db", dbTmp.Close())
				return ok
			})
		})
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) {
		// 打开数据库连接
		db, err := sql.Open(bizConf.GetDriver(), bizConf.GetDsn())
		if !types.IsNil(err) {
			log.Fatalf("Error opening database: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		d.bizDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("close biz db", d.bizDB.Close())
			d.teamBizDBMap.Range(func(key, value any) bool {
				teamBizDB, ok := value.(*gorm.DB)
				if !ok {
					return true
				}
				dbTmp, _ := teamBizDB.DB()
				log.Debugw("close biz orm db", dbTmp.Close())
				return ok
			})
		})
	}

	if !types.IsNil(mainConf) && !types.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn.NewGormDB(mainConf)
		if !types.IsNil(err) {
			cleanup()
			return nil, nil, err
		}
		d.mainDB = mainDB
		// 判断是否有默认用户， 如果没有，则创建一个默认用户
		if err := d.initMainDatabase(); err != nil {
			log.Fatalf("Error init default user: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		// 同步业务模型到各个团队， 保证数据一致性
		if err := d.syncBizDatabase(); err != nil {
			log.Fatalf("Error init biz database: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		closeFuncList = append(closeFuncList, func() {
			mainDBClose, _ := d.mainDB.DB()
			log.Debugw("close main db", mainDBClose.Close())
		})
	}

	return d, cleanup, nil
}

// Exit 推出data
func (d *Data) Exit() <-chan struct{} {
	return d.exit
}

// initMainDatabase 初始化数据库
func (d *Data) initMainDatabase() error {
	if err := d.mainDB.AutoMigrate(model.Models()...); err != nil {
		return err
	}
	// 获取默认用户
	_, err := query.Use(d.mainDB).SysUser.First()
	if err == nil {
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	pass := types.NewPassword("123456")
	encryptValue, err := pass.GetEncryptValue()
	if err != nil {
		return err
	}
	// 如果没有默认用户，则创建一个默认用户
	user := &model.SysUser{
		Username: "admin",
		Nickname: "超级管理员",
		Password: encryptValue,
		Email:    "moonio@moon.com",
		Phone:    "18812341234",
		Remark:   "这是个人很懒， 没有设置备注信息",
		Avatar:   "https://img0.baidu.com/it/u=1128422789,3129806361&fm=253&app=120&size=w931&n=0&f=JPEG&fmt=auto?sec=1719766800&t=ff6081f1e5a590b3033596a43d165f3e",
		Salt:     pass.GetSalt(),
		Gender:   vobj.GenderMale,
		Role:     vobj.RoleSuperAdmin,
		Status:   vobj.StatusEnable,
	}

	return query.Use(d.mainDB).SysUser.Clauses(clause.OnConflict{DoNothing: true}).Create(user)
}

// syncBizDatabase 同步业务模型到各个团队， 保证数据一致性
func (d *Data) syncBizDatabase() error {
	// 获取所有团队
	teams, err := query.Use(d.mainDB).SysTeam.Find()
	if err != nil {
		return err
	}
	for _, team := range teams {
		// 获取团队业务库连接
		db, err := d.GetBizGormDB(team.ID)
		if err != nil {
			return err
		}
		if err := db.AutoMigrate(bizmodel.Models()...); err != nil {
			return err
		}
		// TODO 同步实时告警数据库
		alarmDB, err := d.GetAlarmGormDB(team.ID)
		if err != nil {
			return err
		}
		if err := alarmDB.AutoMigrate(bizmodel.AlarmModels()...); err != nil {
			return err
		}
	}
	return nil
}

// GetMainDB 获取主库连接
func (d *Data) GetMainDB(ctx context.Context) *gorm.DB {
	db, exist := conn.GetDB(ctx)
	if exist {
		return db
	}
	return d.mainDB
}

// GetBizDB 获取业务库连接
func (d *Data) GetBizDB(_ context.Context) *sql.DB {
	return d.bizDB
}

// GenBizDatabaseName 生成业务库名称
func GenBizDatabaseName(teamID uint32) string {
	return fmt.Sprintf("team_%d", teamID)
}

// GetBizGormDB 获取业务库连接
func (d *Data) GetBizGormDB(teamID uint32) (*gorm.DB, error) {
	if teamID == 0 {
		return nil, merr.ErrorI18nNoTeamErr(context.Background())
	}
	dbValue, exist := d.teamBizDBMap.Load(teamID)
	if exist {
		bizDB, isBizDB := dbValue.(*gorm.DB)
		if isBizDB {
			return bizDB, nil
		}
		return nil, merr.ErrorNotification("数据库服务异常")
	}

	dsn := d.bizDatabaseConf.GetDsn() + GenBizDatabaseName(teamID) + "?charset=utf8mb4&parseTime=True&loc=Local"
	bizDbConf := &palaceconf.Data_Database{
		Driver: d.bizDatabaseConf.GetDriver(),
		Dsn:    dsn,
		Debug:  d.bizDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(bizDbConf)
	if !types.IsNil(err) {
		return nil, err
	}

	d.teamBizDBMap.Store(teamID, bizDB)
	enforcer, err := rbac.InitCasbinModel(bizDB)
	if !types.IsNil(err) {
		log.Errorw("casbin init error", err)
		return nil, err
	}
	d.enforcerMap.Store(teamID, enforcer)
	return bizDB, nil
}

// GetAlarmGormDB 获取告警库连接
func (d *Data) GetAlarmGormDB(teamID uint32) (*gorm.DB, error) {
	if teamID == 0 {
		return nil, merr.ErrorI18nNoTeamErr(context.Background())
	}
	dbValue, exist := d.alarmDBMap.Load(teamID)
	if exist {
		bizDB, isBizDB := dbValue.(*gorm.DB)
		if isBizDB {
			return bizDB, nil
		}
		return nil, merr.ErrorNotification("数据库服务异常")
	}

	dsn := d.alarmDatabaseConf.GetDsn() + GenBizDatabaseName(teamID) + "?charset=utf8mb4&parseTime=True&loc=Local"
	alarmDbConf := &palaceconf.Data_Database{
		Driver: d.alarmDatabaseConf.GetDriver(),
		Dsn:    dsn,
		Debug:  d.alarmDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(alarmDbConf)
	if !types.IsNil(err) {
		return nil, err
	}

	d.alarmDBMap.Store(teamID, bizDB)
	return bizDB, nil
}

// GetCacher 获取缓存
func (d *Data) GetCacher() conn.Cache {
	if types.IsNil(d.cacher) {
		log.Warn("cache is nil")
	}
	return d.cacher
}

// GetCasbin 获取casbin
func (d *Data) GetCasbin(teamID uint32) *casbin.SyncedEnforcer {
	enforceVal, exist := d.enforcerMap.Load(teamID)
	if !exist {
		_, err := d.GetBizGormDB(teamID)
		if !types.IsNil(err) {
			panic(err)
		}
	}
	enforce, ok := enforceVal.(*casbin.SyncedEnforcer)
	if !ok {
		panic("enforcer not found")
	}
	return enforce
}

// newCache new cache
func newCache(c *palaceconf.Data_Cache) conn.Cache {
	if types.IsNil(c) {
		return nil
	}

	if !types.IsNil(c.GetRedis()) {
		log.Debugw("cache init", "redis")
		cli := conn.NewRedisClient(c.GetRedis())
		if err := cli.Ping(context.Background()).Err(); !types.IsNil(err) {
			log.Warnw("redis ping error", err)
		}
		return rediscacher.NewRedisCacher(cli)
	}

	if !types.IsNil(c.GetNutsDB()) {
		log.Debugw("cache init", "nutsdb")
		cli, err := nutsdbcacher.NewNutsDbCacher(c.GetNutsDB())
		if !types.IsNil(err) {
			log.Warnw("nutsdb init error", err)
		}
		return cli
	}
	return nil
}
