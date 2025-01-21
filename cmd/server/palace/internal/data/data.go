package data

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/helper/sse"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/oss"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/conn/rbac"
	"github.com/aide-family/moon/pkg/util/email"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData)

// Data .
type Data struct {
	bizDatabaseConf   *conf.Database
	alarmDatabaseConf *conf.Database

	mainDB       *gorm.DB
	alarmDB      *sql.DB
	bizDB        *sql.DB
	cacher       cache.ICacher
	enforcerMap  *safety.Map[uint32, *casbin.SyncedEnforcer]
	teamBizDBMap *safety.Map[uint32, *gorm.DB]
	alarmDBMap   *safety.Map[uint32, *gorm.DB]

	ossClient oss.Client
	ossConf   *conf.Oss

	// 策略队列
	strategyQueue watch.Queue
	// 告警队列
	alertQueue watch.Queue
	// 持久化队列
	alertPersistenceDBQueue watch.Queue
	// 告警持久化存储
	alertConsumerStorage watch.Storage
	// 通用邮件发送器
	emailCli email.Interface
	// sse客户端管理
	sseClientManager *sse.ClientManager
}

var closeFuncList []func()

// NewData .
func NewData(c *palaceconf.Bootstrap) (*Data, func(), error) {
	mainConf := c.GetDatabase()
	alarmConf := c.GetAlarmDatabase()
	bizConf := c.GetBizDatabase()
	cacheConf := c.GetCache()
	emailConf := c.GetEmailConfig()
	ossConf := c.GetOss()
	d := &Data{
		bizDatabaseConf:         bizConf,
		alarmDatabaseConf:       alarmConf,
		teamBizDBMap:            safety.NewMap[uint32, *gorm.DB](),
		alarmDBMap:              safety.NewMap[uint32, *gorm.DB](),
		enforcerMap:             safety.NewMap[uint32, *casbin.SyncedEnforcer](),
		strategyQueue:           watch.NewDefaultQueue(watch.QueueMaxSize),
		alertQueue:              watch.NewDefaultQueue(watch.QueueMaxSize),
		alertPersistenceDBQueue: watch.NewDefaultQueue(watch.QueueMaxSize),
		alertConsumerStorage:    watch.NewDefaultStorage(),
		emailCli:                email.NewMockEmail(),
		ossConf:                 ossConf,
		sseClientManager:        sse.NewClientManager(),
	}
	cleanup := func() {
		d.sseClientManager.Close()
		for _, f := range closeFuncList {
			f()
		}
		log.Info("closing the data resources")
	}
	closeFuncList = append(closeFuncList, func() {
		log.Debug("close data")
	})

	if !types.IsNil(emailConf) {
		d.emailCli = email.New(emailConf)
	}

	d.cacher = cache.NewCache(cacheConf)

	// 是否开启oss
	if ossConf.GetOpen() {
		d.ossClient = newOssCli(ossConf)
	}

	closeFuncList = append(closeFuncList, func() {
		log.Debugw("close cache", d.cacher.Close())
	})

	if !types.IsNil(alarmConf) && !types.TextIsNull(alarmConf.GetDsn()) && alarmConf.GetDriver() != "sqlite" {
		// 打开数据库连接
		db, err := sql.Open(alarmConf.GetDriver(), bizConf.GetDsn())
		if !types.IsNil(err) {
			log.Fatalf("Error opening database: %v\n", err)
			cleanup()
			return nil, nil, err
		}
		d.alarmDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("msg", "close alarm db", "err", d.bizDB.Close())
			for _, value := range d.alarmDBMap.List() {
				dbTmp, _ := value.DB()
				log.Debugw("msg", "close alarm orm db", "err", dbTmp.Close())
			}
		})
	}

	if !types.IsNil(bizConf) && !types.TextIsNull(bizConf.GetDsn()) && bizConf.GetDriver() != "sqlite" {
		// 打开数据库连接
		db, err := sql.Open(bizConf.GetDriver(), bizConf.GetDsn())
		if !types.IsNil(err) {
			log.Fatalf("Error opening database: %v\n", err)
			cleanup()
			return nil, nil, err
		}

		d.bizDB = db
		closeFuncList = append(closeFuncList, func() {
			log.Debugw("msg", "close biz db", "err", d.bizDB.Close())
			for _, value := range d.teamBizDBMap.List() {
				teamBizDB, _ := value.DB()
				log.Debugw("msg", "close biz orm db", "err", teamBizDB.Close())
			}
		})
	}

	if !types.IsNil(mainConf) && !types.TextIsNull(mainConf.GetDsn()) {
		mainDB, err := conn.NewGormDB(mainConf, log.GetLogger())
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

// initMainDatabase 初始化数据库
func (d *Data) initMainDatabase() error {
	return initMainDatabase(d)
}

// syncBizDatabase 同步业务模型到各个团队， 保证数据一致性
func (d *Data) syncBizDatabase() error {
	return syncBizDatabase(d)
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

// CreateBizDatabase 创建业务库
func (d *Data) CreateBizDatabase(teamID uint32) error {
	switch strings.ToLower(d.bizDatabaseConf.GetDriver()) {
	case "mysql":
		ctx := context.Background()
		_, err := d.GetBizDB(ctx).Exec("CREATE DATABASE IF NOT EXISTS " + "`" + genBizDatabaseName(teamID) + "`")
		return err
	default:
		_, err := d.GetBizGormDB(teamID)
		return err
	}
}

// CreateBizAlarmDatabase 创建告警历史业务库
func (d *Data) CreateBizAlarmDatabase(teamID uint32) error {
	switch strings.ToLower(d.alarmDatabaseConf.GetDriver()) {
	case "mysql":
		ctx := context.Background()
		_, err := d.GetBizDB(ctx).Exec("CREATE DATABASE IF NOT EXISTS " + "`" + genAlarmDatabaseName(teamID) + "`")
		return err
	default:
		_, err := d.GetBizGormDB(teamID)
		return err
	}
}

// GetEmail 获取邮件发送器
func (d *Data) GetEmail() email.Interface {
	if types.IsNil(d.emailCli) {
		return email.NewMockEmail()
	}
	return d.emailCli
}

// genBizDatabaseName 生成业务库名称
func genBizDatabaseName(teamID uint32) string {
	return fmt.Sprintf("db_team_%d", teamID)
}

// GenBizDatabaseName 生成业务库名称
func GenBizDatabaseName(teamID uint32) string {
	return genBizDatabaseName(teamID)
}

// genAlarmDatabaseName 生成业务库名称
func genAlarmDatabaseName(teamID uint32) string {
	return fmt.Sprintf("db_team_alarm_%d", teamID)
}

// GetBizGormDBByName 获取业务库连接
func (d *Data) GetBizGormDBByName(databaseName string) (*gorm.DB, error) {
	if databaseName == "" {
		return nil, merr.ErrorNotification("数据库服务异常")
	}
	dsn := databaseName
	driver := strings.ToLower(d.bizDatabaseConf.GetDriver())
	switch driver {
	case "mysql":
		dsn = d.bizDatabaseConf.GetDsn() + databaseName + "?charset=utf8mb4&parseTime=True&loc=Local"
	}

	bizDbConf := &conf.Database{
		Driver: driver,
		Dsn:    dsn,
		Debug:  d.bizDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(bizDbConf, log.GetLogger())
	if !types.IsNil(err) {
		return nil, err
	}
	return bizDB, nil
}

// GetBizGormDB 获取业务库连接
func (d *Data) GetBizGormDB(teamID uint32) (*gorm.DB, error) {
	if teamID == 0 {
		return nil, merr.ErrorI18nToastTeamNotFound(context.Background()).WithMetadata(map[string]string{
			"teamID": strconv.FormatUint(uint64(teamID), 10),
		})
	}
	bizDB, exist := d.teamBizDBMap.Get(teamID)
	if exist {
		return bizDB, nil
	}
	dsn := genBizDatabaseName(teamID)
	driver := strings.ToLower(d.bizDatabaseConf.GetDriver())
	switch driver {
	case "mysql":
		if err := d.CreateBizDatabase(teamID); !types.IsNil(err) {
			return nil, err
		}
		dsn = d.bizDatabaseConf.GetDsn() + genBizDatabaseName(teamID) + "?charset=utf8mb4&parseTime=True&loc=Local"
	}

	bizDbConf := &conf.Database{
		Driver: driver,
		Dsn:    dsn,
		Debug:  d.bizDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(bizDbConf, log.GetLogger())
	if !types.IsNil(err) {
		return nil, err
	}

	d.teamBizDBMap.Set(teamID, bizDB)
	enforcer, err := rbac.InitCasbinModel(bizDB)
	if !types.IsNil(err) {
		log.Errorw("casbin", "init error", "err", err)
		return nil, err
	}
	d.enforcerMap.Set(teamID, enforcer)
	return bizDB, nil
}

// GetAlarmGormDB 获取告警库连接
func (d *Data) GetAlarmGormDB(teamID uint32) (*gorm.DB, error) {
	if teamID == 0 {
		return nil, merr.ErrorI18nToastTeamNotFound(context.Background())
	}
	dbValue, exist := d.alarmDBMap.Get(teamID)
	if exist {
		return dbValue, nil
	}

	dsn := genAlarmDatabaseName(teamID)
	driver := strings.ToLower(d.alarmDatabaseConf.GetDriver())
	switch driver {
	case "mysql":
		if err := d.CreateBizAlarmDatabase(teamID); !types.IsNil(err) {
			return nil, err
		}
		dsn = d.bizDatabaseConf.GetDsn() + genAlarmDatabaseName(teamID) + "?charset=utf8mb4&parseTime=True&loc=Local"
	}

	alarmDbConf := &conf.Database{
		Driver: driver,
		Dsn:    dsn,
		Debug:  d.alarmDatabaseConf.GetDebug(),
	}
	bizDB, err := conn.NewGormDB(alarmDbConf, log.GetLogger())
	if !types.IsNil(err) {
		return nil, err
	}
	d.alarmDBMap.Set(teamID, bizDB)
	return bizDB, nil
}

// GetCacher 获取缓存
func (d *Data) GetCacher() cache.ICacher {
	if types.IsNil(d.cacher) {
		panic("cacher is nil")
	}
	return d.cacher
}

// GetCasbinByTx 获取casbin
func (d *Data) GetCasbinByTx(tx *gorm.DB) *casbin.SyncedEnforcer {
	enforcer, err := rbac.InitCasbinModel(tx)
	if !types.IsNil(err) {
		log.Errorw("casbin", "init error", "err", err)
		panic(err)
	}
	return enforcer
}

// GetCasBin 获取casbin
func (d *Data) GetCasBin(teamID uint32, tx ...*gorm.DB) *casbin.SyncedEnforcer {
	if len(tx) > 0 {
		return d.GetCasbinByTx(tx[0])
	}
	enforceVal, exist := d.enforcerMap.Get(teamID)
	if !exist {
		_, err := d.GetBizGormDB(teamID)
		if !types.IsNil(err) {
			panic(err)
		}
		return d.GetCasBin(teamID)
	}
	return enforceVal
}

// GetStrategyQueue 获取策略队列
func (d *Data) GetStrategyQueue() watch.Queue {
	if types.IsNil(d.strategyQueue) {
		log.Warn("strategyQueue is nil")
	}
	return d.strategyQueue
}

// GetAlertQueue 获取告警队列
func (d *Data) GetAlertQueue() watch.Queue {
	if types.IsNil(d.alertQueue) {
		log.Warn("alertQueue is nil")
	}
	return d.alertQueue
}

// GetAlertPersistenceDBQueue 获取持久化队列
func (d *Data) GetAlertPersistenceDBQueue() watch.Queue {
	if types.IsNil(d.alertPersistenceDBQueue) {
		log.Warn("persistence queue is nil")
	}
	return d.alertPersistenceDBQueue
}

// GetAlertConsumerStorage 获取告警持久化存储
func (d *Data) GetAlertConsumerStorage() watch.Storage {
	if types.IsNil(d.alertConsumerStorage) {
		log.Warn("alertConsumerStorage is nil")
	}
	return d.alertConsumerStorage
}

func newOssCli(c *conf.Oss) oss.Client {
	var client oss.Client
	switch c.GetType() {
	case "aliyun":
		aliOSS, err := oss.NewAliOSS(c.GetAliOss())
		if !types.IsNil(err) {
			panic(err)
		}
		client = aliOSS
	case "tencent":
		tencentOss, err := oss.NewTencentOss(c.GetTencentOss())
		if !types.IsNil(err) {
			panic(err)
		}
		client = tencentOss
	case "minio":
		minIOClient, err := oss.NewMinIO(c.GetMinio())
		if !types.IsNil(err) {
			panic(err)
		}
		client = minIOClient
	case "local":
		client = oss.NewLocalStorage(c.GetLocal())
	default:
		client = oss.NewLocalStorage(c.GetLocal())
	}
	return client
}

// GetOssCli 获取oss客户端
func (d *Data) GetOssCli() oss.Client {
	if types.IsNil(d.ossClient) {
		log.Warn("persistence ossClient is nil")
	}
	return d.ossClient
}

// GetFileLimitSize 获取文件大小限制配置
func (d *Data) GetFileLimitSize() map[string]*conf.FileLimit {
	if types.IsNil(d.ossConf) {
		return map[string]*conf.FileLimit{}
	}
	return d.ossConf.GetLimitSize()
}

// OssIsOpen 是否开启oss
func (d *Data) OssIsOpen() bool {
	if types.IsNil(d.ossConf) {
		return false
	}
	return d.ossConf.GetOpen()
}

// GetSSEClientManager 获取sse客户端管理
func (d *Data) GetSSEClientManager() *sse.ClientManager {
	if types.IsNil(d.sseClientManager) {
		log.Warn("persistence sseClientManager is nil")
		d.sseClientManager = sse.NewClientManager()
	}
	return d.sseClientManager
}
