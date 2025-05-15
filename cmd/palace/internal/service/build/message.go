package build

import (
	"time"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type GetSendMessageLogsRequest interface {
	GetPagination() *common.PaginationRequest
	GetRequestId() string
	GetStatus() common.SendMessageStatus
	GetKeyword() string
	GetMessageType() common.MessageType
}

func ToListSendMessageLogParams(req GetSendMessageLogsRequest) *bo.ListSendMessageLogParams {
	if validate.IsNil(req) {
		panic("GetSendMessageLogsRequest is nil")
	}
	return &bo.ListSendMessageLogParams{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		TeamID:            0,
		RequestID:         req.GetRequestId(),
		Status:            vobj.SendMessageStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
		TimeRange:         [2]time.Time{},
		MessageType:       vobj.MessageType(req.GetMessageType()),
	}
}

func ToGetSendMessageLogParams(requestId string) *bo.GetSendMessageLogParams {
	return &bo.GetSendMessageLogParams{
		TeamID:    0,
		RequestID: requestId,
	}
}

func ToRetrySendMessageParams(requestId string) *bo.RetrySendMessageParams {
	return &bo.RetrySendMessageParams{
		TeamID:    0,
		RequestID: requestId,
	}
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
