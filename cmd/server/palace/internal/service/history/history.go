package service

import (
	"context"

	historyapi "github.com/aide-family/moon/api/admin/history"
)

type HistoryService struct {
	historyapi.UnimplementedHistoryServer
}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

func (s *HistoryService) GetHistory(ctx context.Context, req *historyapi.GetHistoryRequest) (*historyapi.GetHistoryReply, error) {
	return &historyapi.GetHistoryReply{}, nil
}
func (s *HistoryService) ListHistory(ctx context.Context, req *historyapi.ListHistoryRequest) (*historyapi.ListHistoryReply, error) {
	return &historyapi.ListHistoryReply{}, nil
}
