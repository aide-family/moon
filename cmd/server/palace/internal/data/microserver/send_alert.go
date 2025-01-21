package microserver

import (
	"context"
	"strconv"
	"strings"
	"time"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/helper/metric"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/go-kratos/kratos/v2/log"
)

// NewSendAlertRepository 创建告警发送仓库
func NewSendAlertRepository(data *data.Data, rabbitConn *data.RabbitConn, alarmSendRepository repository.AlarmSendRepository) microrepository.SendAlert {
	return &sendAlertRepositoryImpl{rabbitConn: rabbitConn, data: data, alarmSendRepository: alarmSendRepository}
}

// sendAlertRepositoryImpl 告警发送仓库
type sendAlertRepositoryImpl struct {
	data                *data.Data
	rabbitConn          *data.RabbitConn
	alarmSendRepository repository.AlarmSendRepository
}

// Send 发送告警
func (s *sendAlertRepositoryImpl) Send(_ context.Context, alertMsg *bo.SendMsg) {
	go func() {
		defer after.RecoverX()
		s.send(alertMsg)
	}()
}

// send 发送告警
func (s *sendAlertRepositoryImpl) send(task *bo.SendMsg) {
	setOK, err := s.data.GetCacher().Client().SetNX(context.Background(), task.RequestID, "1", 2*time.Hour).Result()
	if err != nil {
		log.Warnf("set cache failed: %v", err)
		return
	}
	if !setOK {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sendStatus := vobj.SentSuccess
	if err := s.rabbitConn.SendMsg(ctx, task.SendMsgRequest); err != nil {
		// 删除缓存
		if err := s.data.GetCacher().Client().Del(context.Background(), task.RequestID).Err(); err != nil {
			log.Warnf("send alert failed")
		}
		sendStatus = vobj.SendFail
		retryNum, err := s.getSendRetryNum(ctx, task.SendMsgRequest)
		// TODO  判断是否重试 默认最大重试次数5次，后续改造可配置
		if err == nil && retryNum > 5 {
			sendStatus = vobj.Sending
			// 加入消息队列，重试
			if err := s.data.GetAlertPersistenceDBQueue().Push(watch.NewMessage(task, vobj.TopicAlertMsg)); err != nil {
				log.Warnf("send alert failed: %v", err)
				sendStatus = vobj.SendFail
			}
		}
	}

	if err = s.alarmSendHistorySave(ctx, task.SendMsgRequest, sendStatus); err != nil {
		log.Error("alarmSendHistorySave failed: ", err)
	}
}

// alarmSendHistorySave 告警发送历史保存
func (s *sendAlertRepositoryImpl) alarmSendHistorySave(ctx context.Context, sendMsg *hookapi.SendMsgRequest, status vobj.SendStatus) error {
	param := builder.NewParamsBuild(ctx).
		AlarmSendModuleBuilder().
		WithCreateAlarmSendRequest(ctx, sendMsg).
		ToBo()

	param.SendStatus = status
	route := sendMsg.Route
	routeParts := strings.Split(route, "_")
	// 检查route是否合法
	if len(routeParts) != 3 {
		log.Warnf("route is invalid: %s", merr.ErrorI18nParameterRelatedAlarmSendingAndReceivingParametersAreInvalid(ctx))
		return nil
	}
	teamID := routeParts[1]
	notifyID := routeParts[2]
	notifyIDInt, _ := strconv.Atoi(notifyID)
	teamIDInt, _ := strconv.Atoi(teamID)
	strategyNameMap := s.getStrategyNameMap(uint32(teamIDInt), []uint32{uint32(notifyIDInt)})
	metric.IncNotifyCounter(teamID, status.EnString(), notifyID, strategyNameMap[uint32(notifyIDInt)])
	// 解析告警team_id
	param.TeamID = getTeamIDByRoute(routeParts)
	// 解析告警组id
	param.AlarmGroupID = getAlarmGroupIDByRoute(routeParts)
	param.SendTime = types.NewTime(time.Now())
	return s.alarmSendRepository.SaveAlarmSendHistory(ctx, param)
}

// getBizQuery 获取业务数据库
func getBizQuery(ctx context.Context, data *data.Data) (*bizquery.Query, error) {
	bizDB, err := data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return nil, err
	}
	return bizquery.Use(bizDB), nil
}

// getStrategyNameMap 获取策略名称
func (s *sendAlertRepositoryImpl) getStrategyNameMap(teamID uint32, strategyIDList []uint32) map[uint32]string {
	if len(strategyIDList) == 0 {
		return nil
	}
	ctx := middleware.WithTeamIDContextKey(context.Background(), teamID)
	bizQuery, err := getBizQuery(ctx, s.data)
	if err != nil {
		return nil
	}
	strategies, err := bizQuery.Strategy.WithContext(ctx).
		Unscoped().
		Select(bizQuery.Strategy.ID, bizQuery.Strategy.Name).
		Where(bizQuery.Strategy.ID.In(strategyIDList...)).
		Find()
	if err != nil {
		return nil
	}
	strategyNameMap := make(map[uint32]string)
	for _, strategy := range strategies {
		strategyNameMap[strategy.ID] = strategy.Name
	}
	return strategyNameMap
}

// 获取告警发送次数
func (s *sendAlertRepositoryImpl) getSendRetryNum(ctx context.Context, sendMsg *hookapi.SendMsgRequest) (int, error) {
	route := sendMsg.Route
	requestID := sendMsg.RequestID

	routeParts := strings.Split(route, "_")
	// 检查route是否合法
	if len(routeParts) != 3 {
		return 0, merr.ErrorI18nParameterRelatedAlarmSendingAndReceivingParametersAreInvalid(ctx)
	}

	teamID := getTeamIDByRoute(routeParts)
	return s.alarmSendRepository.GetRetryNumberByRequestID(ctx, requestID, teamID)
}

// 解析告警team id
func getTeamIDByRoute(routeParts []string) uint32 {
	teamID, _ := strconv.ParseInt(routeParts[1], 10, 32)
	return uint32(teamID)
}

// 解析告警组id
func getAlarmGroupIDByRoute(routeParts []string) uint32 {
	teamID, _ := strconv.ParseInt(routeParts[2], 10, 32)
	return uint32(teamID)
}
