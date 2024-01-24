package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"

	"prometheus-manager/api"
	alarmhookPb "prometheus-manager/api/alarm/hook"
	"prometheus-manager/api/prom/strategy/group"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/servers"
	"prometheus-manager/pkg/util/cache"
)

var _ transport.Server = (*AlarmEvent)(nil)

type EventHandler func(msg *kafka.Message) error

type AgentInfo struct {
	AgentName string
	Topic     *string
	Key       []byte
}

// String AgentInfo to string
func (a *AgentInfo) String() string {
	bs, _ := json.Marshal(a)
	return string(bs)
}

type AlarmEvent struct {
	log    *log.Helper
	c      *conf.Bootstrap
	exitCh chan struct{}

	kafkaMQServer      *servers.KafkaMQServer
	hookService        *alarmservice.HookService
	groupService       *promservice.GroupService
	changeGroupChannel <-chan uint32
	removeGroupChannel <-chan bo.RemoveStrategyGroupBO

	agentNames     cache.Cache
	groups         cache.Cache
	changeGroupIds cache.Cache
	eventHandlers  map[consts.TopicType]EventHandler
}

func (l *AlarmEvent) Start(_ context.Context) error {
	kafkaConf := l.c.GetMq().GetKafka()
	if err := l.storeGroups(); err != nil {
		return err
	}

	if err := l.Subscribe(kafkaConf.GetTopics()); err != nil {
		return err
	}

	if err := l.Consume(); err != nil {
		return err
	}

	if err := l.watchChangeGroup(); err != nil {
		return err
	}

	return nil
}

func (l *AlarmEvent) Stop(_ context.Context) error {
	l.log.Info("AlarmEvent stopping")
	l.kafkaMQServer.Producer().Close()
	if err := l.kafkaMQServer.Consumer().Close(); err != nil {
		return err
	}

	close(l.exitCh)
	l.log.Info("AlarmEvent stopped")
	return nil
}

func NewAlarmEvent(
	c *conf.Bootstrap,
	d *data.Data,
	changeGroupChannel <-chan uint32,
	removeGroupChannel <-chan bo.RemoveStrategyGroupBO,
	hookService *alarmservice.HookService,
	groupService *promservice.GroupService,
	logger log.Logger,
) (*AlarmEvent, error) {
	kafkaConf := c.GetMq().GetKafka()
	kafkaMqServer, err := servers.NewKafkaMQServer(kafkaConf, logger)
	if err != nil {
		return nil, err
	}

	l := &AlarmEvent{
		log:                log.NewHelper(log.With(logger, "module", "server.alarm.event")),
		c:                  c,
		exitCh:             make(chan struct{}),
		eventHandlers:      make(map[consts.TopicType]EventHandler),
		kafkaMQServer:      kafkaMqServer,
		hookService:        hookService,
		groupService:       groupService,
		changeGroupChannel: changeGroupChannel,
		removeGroupChannel: removeGroupChannel,
		agentNames:         cache.NewRedisCache(d.Client(), consts.AgentNames),
		groups:             cache.NewRedisCache(d.Client(), consts.StrategyGroups),
		changeGroupIds:     cache.NewRedisCache(d.Client(), consts.ChangeGroupIds),
	}

	// 注册topic处理器
	for _, topic := range kafkaConf.GetTopics() {
		switch topic {
		case string(consts.AlertHookTopic):
			l.eventHandlers[consts.AlertHookTopic] = l.alertHookHandler
		case string(consts.AgentOnlineTopic):
			l.eventHandlers[consts.AgentOnlineTopic] = l.agentOnlineEventHandler
		case string(consts.AgentOfflineTopic):
			l.eventHandlers[consts.AgentOfflineTopic] = l.agentOfflineEventHandler
		default:
			return nil, fmt.Errorf("topic %s not support", topic)
		}
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
		l.groups.Store(groupIdStr, string(groupItemBytes))
	}

	return nil
}

func (l *AlarmEvent) watchChangeGroup() error {
	// 一分钟执行一次
	//ticker := time.NewTicker(time.Minute * 1)
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		defer after.Recover(l.log)
		for {
			select {
			case <-l.exitCh:
				return
			case groupId := <-l.changeGroupChannel:
				groupIdStr := strconv.FormatUint(uint64(groupId), 10)
				l.changeGroupIds.Store(groupIdStr, "")
			case groupInfo := <-l.removeGroupChannel:
				if err := l.sendRemoveGroup(groupInfo.Id); err != nil {
					l.log.Errorf("send remove group error: %s", err.Error())
				}
			case <-ticker.C:
				l.log.Infof("start synce store groups")
				changeGroupIds := make([]uint32, 0, 128)
				l.changeGroupIds.Range(func(key, value string) bool {
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

				l.log.Infof("synce store groups success %v", len(changeGroupIds))

				for _, groupItem := range listAllGroupDetail.GetGroupList() {
					groupIdStr := strconv.FormatUint(uint64(groupItem.GetId()), 10)
					groupItemBytes, err := json.Marshal(groupItem)
					if err != nil {
						continue
					}
					l.groups.Store(groupIdStr, string(groupItemBytes))
					if err = l.sendChangeGroup(groupItem); err != nil {
						l.log.Errorf("send change group error: %s", err.Error())
					}
					l.changeGroupIds.Delete(groupIdStr)
				}
			}
		}
	}()
	return nil
}

func (l *AlarmEvent) handleMessage(msg *kafka.Message) bool {
	topic := *msg.TopicPartition.Topic
	l.log.Infow("topic", topic)
	if handler, ok := l.eventHandlers[consts.TopicType(topic)]; ok {
		if err := handler(msg); err != nil {
			l.log.Errorf("handle message error: %s", err.Error())
		}
	}
	return true
}

// alertHook 处理alert hook数据
func (l *AlarmEvent) alertHookHandler(msg *kafka.Message) error {
	var req alarmhookPb.HookV1Request
	// TODO 后期是否判断key
	err := json.Unmarshal(msg.Value, &req)
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
func (l *AlarmEvent) agentOfflineEventHandler(msg *kafka.Message) error {
	l.log.Infof("agent offline: %s", string(msg.Key))
	l.agentNames.Delete(string(msg.Key))
	return nil
}

// agentOnlineEventHandler 处理agent online消息
func (l *AlarmEvent) agentOnlineEventHandler(msg *kafka.Message) error {
	// 记录节点状态
	topic := string(msg.Value)
	l.log.Infof("agent online: %s, topic: %s", string(msg.Key), topic)
	agentInfo := &AgentInfo{
		AgentName: string(msg.Key),
		Topic:     &topic,
		Key:       msg.Key,
	}

	l.agentNames.Store(string(msg.Key), agentInfo.String())

	eg := new(errgroup.Group)
	eg.SetLimit(100)
	l.groups.Range(func(key, groupDetail string) bool {
		eg.Go(func() error {
			// 3. 推送规则组消息(按规则组粒度)
			sendMsg := &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     agentInfo.Topic,
					Partition: kafka.PartitionAny,
				},
				Value: []byte(groupDetail),
				Key:   agentInfo.Key,
			}
			return l.kafkaMQServer.Produce(sendMsg)
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
	l.agentNames.Range(func(key, agentInfoStr string) bool {
		var agentInfo AgentInfo
		if err := json.Unmarshal([]byte(agentInfoStr), &agentInfo); err != nil {
			return true
		}
		eg.Go(func() error {
			// 3. 推送规则组消息(按规则组粒度)
			sendMsg := &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Value: groupDetailBytes,
				Key:   agentInfo.Key,
			}
			return l.kafkaMQServer.Produce(sendMsg)
		})
		return true
	})

	return eg.Wait()
}

// sendRemoveGroup 发送移除规则组消息
func (l *AlarmEvent) sendRemoveGroup(groupId uint32) error {
	groupIdStr := strconv.FormatUint(uint64(groupId), 10)
	l.groups.Delete(groupIdStr)
	l.log.Infof("send remove group: %d", groupId)
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	topic := string(consts.RemoveGroupTopic)
	msgValue := []byte(strconv.FormatUint(uint64(groupId), 10))
	l.agentNames.Range(func(key, agentInfoStr string) bool {
		var agentInfo AgentInfo
		if err := json.Unmarshal([]byte(agentInfoStr), &agentInfo); err != nil {
			return true
		}
		eg.Go(func() error {
			// 3. 推送规则组消息(按规则组粒度)
			sendMsg := &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Value: msgValue,
				Key:   agentInfo.Key,
			}
			return l.kafkaMQServer.Produce(sendMsg)
		})
		return true
	})

	return eg.Wait()
}

// Subscribe 订阅消息
func (l *AlarmEvent) Subscribe(topics []string) error {
	return l.kafkaMQServer.Consume(topics, l.handleMessage)
}

// Consume 消费消息
func (l *AlarmEvent) Consume() error {
	consumer := l.kafkaMQServer.Consumer()
	go func() {
		events := consumer.Events()
		for event := range events {
			if consumer.IsClosed() {
				l.log.Warnf("consumer is closed")
				return
			}
			switch e := event.(type) {
			case *kafka.Message:
				// 处理消息, 根据不同的topic做不同的处理
				l.log.Debugf("Message on %s\n", e.TopicPartition)
				if e.TopicPartition.Topic == nil {
					break
				}
				topic := consts.TopicType(*e.TopicPartition.Topic)
				handler, ok := l.eventHandlers[topic]
				if !ok {
					l.log.Warnf("no handler found for topic: %s, event: %v", topic, e)
					break
				}
				if err := handler(e); err != nil {
					l.log.Errorf("handle event error: %v", err)
				}
			case kafka.Error:
				//l.log.Warnf("Error: %v\n", e)
			}
		}
	}()
	return nil
}
