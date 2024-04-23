package server

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/aide-family/moon/api"
	alarmhookPb "github.com/aide-family/moon/api/alarm/hook"
	"github.com/aide-family/moon/api/server/prom/strategy/group"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/app/prom_server/internal/service/alarmservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/promservice"
	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/util/cache"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"
)

var _ transport.Server = (*AlarmEvent)(nil)

const timeout = 10 * time.Second

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

	interflowInstance interflow.ServerInterflow
}

func (l *AlarmEvent) Start(_ context.Context) error {
	l.log.Debug("[AlarmEvent] starting")
	defer l.log.Debug("[AlarmEvent] started")
	// 通知agent，server已经上线
	agentUrls := make([]string, 0, 10)
	l.agentNameCache.Range(func(key, agentInfoStr string) bool {
		var agentInfo AgentInfo
		if err := json.Unmarshal([]byte(agentInfoStr), &agentInfo); err != nil {
			return true
		}
		agentUrls = append(agentUrls, agentInfo.Url)
		return true
	})

	if err := l.interflowInstance.ServerOnlineNotify(agentUrls); err != nil {
		return err
	}

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
	l.log.Debug("[AlarmEvent] stopping")
	defer l.log.Debug("[AlarmEvent] stopped")
	// 通知agent，server已经离线
	agentUrls := make([]string, 0, 10)
	l.agentNameCache.Range(func(key, agentInfoStr string) bool {
		var agentInfo AgentInfo
		if err := json.Unmarshal([]byte(agentInfoStr), &agentInfo); err != nil {
			return true
		}
		agentUrls = append(agentUrls, agentInfo.Url)
		return true
	})

	if err := l.interflowInstance.ServerOfflineNotify(agentUrls); err != nil {
		return err
	}

	close(l.exitCh)

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

	l.changeGroupIdCache.Range(func(key, value string) bool {
		l.changeGroupIdCache.Delete(key)
		return true
	})

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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 2. 拉取全量规则组及规则
	listAllGroupDetail, err := l.groupService.ListAllGroupDetail(ctx, &group.ListAllGroupDetailRequest{})
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
	ticker := time.NewTicker(time.Second * 10)
	count := 0
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
					l.log.Errorw("send remove group error", err)
				}
			case <-ticker.C:
				//l.log.Debug("start sync store groups")
				changeGroupIds := make([]uint32, 0, 128)
				l.changeGroupIdCache.Range(func(key, value string) bool {
					groupId, err := strconv.ParseUint(key, 10, 64)
					if err != nil {
						l.log.Errorw("parse group id error", err)
						return true
					}
					changeGroupIds = append(changeGroupIds, uint32(groupId))
					return true
				})

				if len(changeGroupIds) == 0 && count > 0 {
					//l.log.Debug("no change group")
					continue
				}
				l.log.Debugw("changeGroupIds", changeGroupIds)
				// 重新拉取全量规则组及规则
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				listAllGroupDetail, err := l.groupService.ListAllGroupDetail(ctx, &group.ListAllGroupDetailRequest{
					GroupIds: changeGroupIds,
				})
				cancel()
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
						l.log.Errorw("send change group error", err)
					}
					l.changeGroupIdCache.Delete(groupIdStr)
				}
				l.log.Debugw("sync store groups done", changeGroupIds)
				count++
			}
		}
	}()
	return nil
}

// alertHook 处理alert hook数据
func (l *AlarmEvent) alertHookHandler(topic consts.TopicType, value []byte) error {
	var req alarmhookPb.HookV1Request
	// TODO 后期是否判断topic
	err := json.Unmarshal(value, &req)
	if err != nil {
		return err
	}
	if err = req.ValidateAll(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := l.hookService.V1(ctx, &req)
	if err != nil {
		return err
	}
	l.log.Debugf("hook resp: %s", resp.String())
	return nil
}

// agentOfflineEventHandler 处理agent offline消息
func (l *AlarmEvent) agentOfflineEventHandler(topic consts.TopicType, value []byte) error {
	l.log.Infof("agent offline: %s, topic: %s", string(value), topic)
	l.agentNameCache.Delete(string(value))
	return nil
}

// agentOnlineEventHandler 处理agent online消息
func (l *AlarmEvent) agentOnlineEventHandler(topic consts.TopicType, value []byte) error {
	// 记录节点状态
	l.log.Infof("agent online: %s, topic: %s", string(value), topic)
	agentInfo := &AgentInfo{
		Topic: string(topic),
		Url:   string(value),
	}

	l.agentNameCache.Store(string(value), agentInfo.String())

	eg := new(errgroup.Group)
	eg.SetLimit(100)
	l.groupCache.Range(func(key, groupDetail string) bool {
		eg.Go(func() error {
			msg := &interflow.HookMsg{
				Topic: string(consts.StrategyGroupAllTopic),
				Value: []byte(groupDetail),
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return l.interflowInstance.SendAgent(ctx, agentInfo.Url, msg)
		})
		return true
	})

	return eg.Wait()
}

// 发送策略组信息
func (l *AlarmEvent) sendChangeGroup(groupDetail *api.GroupSimple) error {
	l.log.Debugw("send change group", groupDetail.Id)
	groupDetailBytes, _ := json.Marshal(groupDetail)
	//l.log.Debugw("groupDetailBytes", string(groupDetailBytes), "groupDetail", groupDetail)
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
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if err := l.interflowInstance.SendAgent(ctx, agentInfo.Url, msg); err != nil {
				l.log.Errorw("send change group error", err)
				// 移除该agent
				l.agentNameCache.Delete(agentInfo.Url)
				return err
			}
			return nil
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
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return l.interflowInstance.SendAgent(ctx, agentInfo.Url, msg)
		})
		return true
	})

	return eg.Wait()
}

type AgentInfo struct {
	Topic string `json:"topic"`
	Url   string `json:"key"`
}

// String AgentInfo to string
func (a *AgentInfo) String() string {
	bs, _ := json.Marshal(a)
	return string(bs)
}
