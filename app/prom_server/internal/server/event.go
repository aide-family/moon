package server

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/api"
	alarmhookPb "prometheus-manager/api/alarm/hook"
	"prometheus-manager/api/server/prom/strategy/group"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/cache"
	"prometheus-manager/pkg/util/interflow"
)

var _ transport.Server = (*AlarmEvent)(nil)

type AlarmEvent struct {
	log    *log.Helper
	exitCh chan struct{}

	d                  *data.Data
	hookService        *alarmservice.HookService
	groupService       *promservice.GroupService
	changeGroupChannel <-chan uint32
	removeGroupChannel <-chan bo.RemoveStrategyGroupBO

	agentNameCache     cache.Cache
	groupCache         cache.Cache
	changeGroupIdCache cache.Cache
	eventHandlers      map[consts.TopicType]interflow.Callback

	interflowInstance interflow.Interflow
}

func (l *AlarmEvent) Start(_ context.Context) error {
	if err := l.storeGroups(); err != nil {
		return err
	}

	if err := l.interflowInstance.Receive(); err != nil {
		return err
	}

	if err := l.watchChangeGroup(); err != nil {
		return err
	}

	return nil
}

func (l *AlarmEvent) Stop(_ context.Context) error {
	l.log.Info("AlarmEvent stopping")
	close(l.exitCh)
	l.log.Info("AlarmEvent stopped")
	return nil
}

func NewAlarmEvent(
	d *data.Data,
	changeGroupChannel <-chan uint32,
	removeGroupChannel <-chan bo.RemoveStrategyGroupBO,
	hookService *alarmservice.HookService,
	groupService *promservice.GroupService,
	logger log.Logger,
) (*AlarmEvent, error) {
	globalCache := d.Cache()
	l := &AlarmEvent{
		log:                log.NewHelper(log.With(logger, "module", "server.alarm.event")),
		exitCh:             make(chan struct{}),
		eventHandlers:      make(map[consts.TopicType]interflow.Callback),
		hookService:        hookService,
		groupService:       groupService,
		changeGroupChannel: changeGroupChannel,
		removeGroupChannel: removeGroupChannel,
		agentNameCache:     cache.NewRedisCache(globalCache, consts.AgentNames),
		groupCache:         cache.NewRedisCache(globalCache, consts.StrategyGroups),
		changeGroupIdCache: cache.NewRedisCache(globalCache, consts.ChangeGroupIds),
		interflowInstance:  d.Interflow(),
	}

	// 注册topic处理器
	l.eventHandlers[consts.AlertHookTopic] = l.alertHookHandler
	l.eventHandlers[consts.AgentOnlineTopic] = l.agentOnlineEventHandler
	l.eventHandlers[consts.AgentOfflineTopic] = l.agentOfflineEventHandler

	if err := l.interflowInstance.SetHandles(l.eventHandlers); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *AlarmEvent) storeGroups() error {
	// 2. 拉取全量规则组及规则
	listAllGroupDetail, err := l.groupService.ListAllGroupDetail(context.Background(), &group.ListAllGroupDetailRequest{})
	if err != nil {
		l.log.Errorf("list all group detail error: %v", err)
		return err
	}
	for _, groupItem := range listAllGroupDetail.GetGroupList() {
		groupIdStr := strconv.FormatUint(uint64(groupItem.GetId()), 10)
		groupItemBytes, err := json.Marshal(groupItem)
		if err != nil {
			continue
		}
		l.groupCache.Store(groupIdStr, string(groupItemBytes))
	}

	return nil
}

func (l *AlarmEvent) watchChangeGroup() error {
	// 一分钟执行一次
	//ticker := time.NewTicker(time.Minute * 10)
	ticker := time.NewTicker(time.Second * 10 * 60)
	go func() {
		defer after.Recover(l.log)
		for {
			select {
			case <-l.exitCh:
				ticker.Stop()
				return
			case groupId := <-l.changeGroupChannel:
				groupIdStr := strconv.FormatUint(uint64(groupId), 10)
				l.changeGroupIdCache.Store(groupIdStr, "")
			case groupInfo := <-l.removeGroupChannel:
				if err := l.sendRemoveGroup(groupInfo.Id); err != nil {
					l.log.Errorf("send remove group error: %s", err.Error())
				}
			case <-ticker.C:
				l.log.Infof("start synce store groups")
				changeGroupIds := make([]uint32, 0, 128)
				l.changeGroupIdCache.Range(func(key, value string) bool {
					groupId, err := strconv.ParseUint(key, 10, 64)
					if err != nil {
						l.log.Errorf("parse group id error: %v", err)
						return true
					}
					changeGroupIds = append(changeGroupIds, uint32(groupId))
					return true
				})

				if len(changeGroupIds) == 0 {
					l.log.Info("no change group")
					continue
				}
				l.log.Infow("changeGroupIds", changeGroupIds)
				// 重新拉取全量规则组及规则
				listAllGroupDetail, err := l.groupService.ListAllGroupDetail(context.Background(), &group.ListAllGroupDetailRequest{
					GroupIds: changeGroupIds,
				})
				if err != nil {
					l.log.Errorf("list all group detail error: %v", err)
					continue
				}
				for _, groupId := range changeGroupIds {
					l.changeGroupIdCache.Delete(strconv.FormatUint(uint64(groupId), 10))
				}

				for _, groupItem := range listAllGroupDetail.GetGroupList() {
					groupIdStr := strconv.FormatUint(uint64(groupItem.GetId()), 10)
					groupItemBytes, err := json.Marshal(groupItem)
					if err != nil {
						continue
					}
					l.groupCache.Store(groupIdStr, string(groupItemBytes))
					if err = l.sendChangeGroup(groupItem); err != nil {
						l.log.Errorf("send change group error: %s", err.Error())
					}
					l.changeGroupIdCache.Delete(groupIdStr)
				}
				l.log.Infof("synce store groups done %v", len(changeGroupIds))
			}
		}
	}()
	return nil
}

// alertHook 处理alert hook数据
func (l *AlarmEvent) alertHookHandler(topic consts.TopicType, key, value []byte) error {
	var req alarmhookPb.HookV1Request
	// TODO 后期是否判断key
	err := json.Unmarshal(value, &req)
	if err != nil {
		return err
	}
	if err = req.ValidateAll(); err != nil {
		return err
	}
	resp, err := l.hookService.V1(context.Background(), &req)
	if err != nil {
		return err
	}
	l.log.Debugf("hook resp: %s", resp.String())
	return nil
}

// agentOfflineEventHandler 处理agent offline消息
func (l *AlarmEvent) agentOfflineEventHandler(topic consts.TopicType, key, value []byte) error {
	l.log.Infof("agent offline: %s", string(key))
	l.agentNameCache.Delete(string(key))
	return nil
}

// agentOnlineEventHandler 处理agent online消息
func (l *AlarmEvent) agentOnlineEventHandler(topic consts.TopicType, key, value []byte) error {
	// 记录节点状态
	sendTopic := string(value)
	l.log.Infof("agent online: %s, topic: %s", string(key), topic)
	agentInfo := &AgentInfo{
		Topic: &sendTopic,
		Key:   key,
	}

	l.agentNameCache.Store(string(key), agentInfo.String())

	eg := new(errgroup.Group)
	eg.SetLimit(100)
	l.groupCache.Range(func(key, groupDetail string) bool {
		eg.Go(func() error {
			msg := &interflow.HookMsg{
				Topic: string(consts.StrategyGroupAllTopic),
				Value: []byte(groupDetail),
				Key:   agentInfo.Key,
			}
			return l.interflowInstance.Send(context.Background(), string(agentInfo.Key), msg)
		})
		return true
	})

	return eg.Wait()
}

// 发送策略组信息
func (l *AlarmEvent) sendChangeGroup(groupDetail *api.GroupSimple) error {
	l.log.Infof("send change group: %d", groupDetail.Id)
	groupDetailBytes, _ := json.Marshal(groupDetail)
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	topic := string(consts.StrategyGroupAllTopic)
	l.agentNameCache.Range(func(key, agentInfoStr string) bool {
		var agentInfo AgentInfo
		if err := json.Unmarshal([]byte(agentInfoStr), &agentInfo); err != nil {
			return true
		}
		eg.Go(func() error {
			msg := &interflow.HookMsg{
				Topic: topic,
				Value: groupDetailBytes,
				Key:   agentInfo.Key,
			}
			return l.interflowInstance.Send(context.Background(), string(agentInfo.Key), msg)
		})
		return true
	})

	return eg.Wait()
}

// sendRemoveGroup 发送移除规则组消息
func (l *AlarmEvent) sendRemoveGroup(groupId uint32) error {
	groupIdStr := strconv.FormatUint(uint64(groupId), 10)
	l.groupCache.Delete(groupIdStr)
	l.log.Infof("send remove group: %d", groupId)
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	topic := string(consts.RemoveGroupTopic)
	msgValue := []byte(strconv.FormatUint(uint64(groupId), 10))
	l.agentNameCache.Range(func(key, agentInfoStr string) bool {
		var agentInfo AgentInfo
		if err := json.Unmarshal([]byte(agentInfoStr), &agentInfo); err != nil {
			return true
		}
		eg.Go(func() error {
			msg := &interflow.HookMsg{
				Topic: topic,
				Value: msgValue,
				Key:   agentInfo.Key,
			}
			return l.interflowInstance.Send(context.Background(), string(agentInfo.Key), msg)
		})
		return true
	})

	return eg.Wait()
}

type AgentInfo struct {
	Topic *string `json:"topic"`
	Key   []byte  `json:"key"`
}

// String AgentInfo to string
func (a *AgentInfo) String() string {
	bs, _ := json.Marshal(a)
	return string(bs)
}
