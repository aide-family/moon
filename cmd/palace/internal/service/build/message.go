package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

type GetSendMessageLogsRequest interface {
	GetPagination() *common.PaginationRequest
	GetRequestId() string
	GetStatus() common.SendMessageStatus
	GetKeyword() string
	GetMessageType() common.MessageType
	GetTimeRange() []string
}

func ToListSendMessageLogParams(req GetSendMessageLogsRequest) (*bo.ListSendMessageLogParams, error) {
	if validate.IsNil(req) {
		panic("GetSendMessageLogsRequest is nil")
	}
	timeRange, err := ToTimeRange(req.GetTimeRange())
	if err != nil {
		return nil, err
	}
	return &bo.ListSendMessageLogParams{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		TeamID:            0,
		RequestID:         req.GetRequestId(),
		Status:            vobj.SendMessageStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
		TimeRange:         timeRange,
		MessageType:       vobj.MessageType(req.GetMessageType()),
	}, nil
}

func ToGetSendMessageLogParams(requestId string) *bo.GetSendMessageLogParams {
	return &bo.GetSendMessageLogParams{
		TeamID:    0,
		RequestID: requestId,
	}
}

type OperateOneSendMessageRequest interface {
	GetRequestId() string
	GetSendTime() string
}

func ToRetrySendMessageParams(req OperateOneSendMessageRequest) (*bo.RetrySendMessageParams, error) {
	sendAt, err := timex.Parse(req.GetSendTime())
	if err != nil {
		return nil, err
	}
	return &bo.RetrySendMessageParams{
		TeamID:    0,
		RequestID: req.GetRequestId(),
		SendAt:    sendAt,
	}, nil
}

func ToUpdateSendMessageLogStatusParams(req *palace.SendMsgCallbackRequest) *bo.UpdateSendMessageLogStatusParams {
	item := &bo.UpdateSendMessageLogStatusParams{
		TeamID:    req.GetTeamId(),
		RequestID: req.GetRequestId(),
	}
	if req.GetCode() == 0 {
		item.Status = vobj.SendMessageStatusSuccess
	} else {
		item.Status = vobj.SendMessageStatusFailed
		item.Error = req.GetMsg()
	}
	return item
}

func ToSendMessageLog(logDo do.SendMessageLog) *common.SendMessageLogItem {
	if validate.IsNil(logDo) {
		return nil
	}
	return &common.SendMessageLogItem{
		RequestId:   logDo.GetRequestID(),
		Message:     logDo.GetMessage(),
		MessageType: common.MessageType(logDo.GetMessageType().GetValue()),
		Status:      common.SendMessageStatus(logDo.GetStatus().GetValue()),
		Error:       logDo.GetError(),
		RetryCount:  logDo.GetRetryCount(),
		CreatedAt:   timex.Format(logDo.GetCreatedAt()),
		UpdatedAt:   timex.Format(logDo.GetUpdatedAt()),
	}
}

func ToSendMessageLogs(logs []do.SendMessageLog) []*common.SendMessageLogItem {
	return slices.Map(logs, ToSendMessageLog)
}
