package bo

import (
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/encoding"
	"github.com/aide-family/magicbox/encoding/json"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/types/known/anypb"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

type CreateMessageLogBo struct {
	Message     strutil.EncryptString
	Config      strutil.EncryptString
	MessageType enum.MessageType
}

func NewCreateMessageLogBo(message, config strutil.EncryptString, messageType enum.MessageType) *CreateMessageLogBo {
	return &CreateMessageLogBo{
		Message:     message,
		Config:      config,
		MessageType: messageType,
	}
}

type MessageLogItemBo struct {
	UID          snowflake.ID
	NamespaceUID snowflake.ID
	SendAt       time.Time
	Message      strutil.EncryptString
	Config       strutil.EncryptString
	MessageType  enum.MessageType
	Status       enum.MessageStatus
	RetryTotal   int32
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (b *MessageLogItemBo) ToAPIV1MessageLogItem() *apiv1.MessageLogItem {
	return &apiv1.MessageLogItem{
		Uid:         b.UID.Int64(),
		MessageType: b.MessageType,
		Status:      enum.MessageStatus(b.Status),
		SendAt:      b.SendAt.Format(time.DateTime),
		Message:     string(b.Message),
		Config:      string(b.Config),
		RetryTotal:  b.RetryTotal,
		LastError:   b.LastError,
		CreatedAt:   b.CreatedAt.Format(time.DateTime),
		UpdatedAt:   b.UpdatedAt.Format(time.DateTime),
	}
}

// ToMessageConfig 将 bo.MessageLogItemBo 的 Config（BO 的 JSON）转换为 *config.MessageConfig。
// 当前 Config 存储的是各业务 BO 的 JSON（如 WebhookItemBo、EmailConfigItemBo），按 messageType 反序列化后
// 组装为 Driver 所需的 MessageConfig（MessageType + Options *anypb.Any）。
func (b *MessageLogItemBo) ToMessageConfig() (*config.MessageConfig, error) {
	configBytes := []byte(string(b.Config))
	msgType := b.MessageType
	jsonCodec, ok := encoding.GetCodec(json.Name)
	if !ok {
		return nil, merr.ErrorInternalServer("%s codec not found", json.Name)
	}
	switch {
	case msgType == enum.MessageType_EMAIL:
		var messageEmailConfig config.MessageEmailConfig
		if err := jsonCodec.Unmarshal(configBytes, &messageEmailConfig); err != nil {
			return nil, merr.ErrorInternalServer("unmarshal email config failed: %v", err)
		}
		options, err := anypb.New(&messageEmailConfig)
		if err != nil {
			return nil, err
		}
		return &config.MessageConfig{MessageType: msgType, Options: options}, nil
	case msgType >= enum.MessageType_WEBHOOK_OTHER && msgType < 3000:
		var messageWebhookConfig config.MessageWebhookConfig
		if err := jsonCodec.Unmarshal(configBytes, &messageWebhookConfig); err != nil {
			return nil, merr.ErrorInternalServer("unmarshal webhook config failed: %v", err)
		}
		options, err := anypb.New(&messageWebhookConfig)
		if err != nil {
			return nil, err
		}
		return &config.MessageConfig{MessageType: msgType, Options: options}, nil
	default:
		return nil, merr.ErrorInternalServer("unsupported message type for config conversion: %s", msgType)
	}
}

type ListMessageLogBo struct {
	*PageRequestBo
	StartAt     time.Time
	EndAt       time.Time
	Status      enum.MessageStatus
	MessageType enum.MessageType
}

func NewListMessageLogBo(req *apiv1.ListMessageLogRequest) *ListMessageLogBo {
	return &ListMessageLogBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		StartAt:       time.Unix(req.StartAtUnix, 0),
		EndAt:         time.Unix(req.EndAtUnix, 0),
		Status:        req.Status,
		MessageType:   req.MessageType,
	}
}

func ToAPIV1ListMessageLogReply(pageResponseBo *PageResponseBo[*MessageLogItemBo]) *apiv1.ListMessageLogReply {
	items := make([]*apiv1.MessageLogItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1MessageLogItem())
	}
	return &apiv1.ListMessageLogReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}
