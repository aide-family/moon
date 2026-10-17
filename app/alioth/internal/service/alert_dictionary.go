package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertdictionary"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type AlertDictionaryService struct {
	v1.UnimplementedAlertDictionaryServiceServer
	dictionaries *alertdictionary.Usecase
}

func NewAlertDictionaryService(uc *alertdictionary.Usecase) *AlertDictionaryService {
	return &AlertDictionaryService{dictionaries: uc}
}

func (s *AlertDictionaryService) CreateAlertDictionary(ctx context.Context, req *v1.CreateAlertDictionaryRequest) (*v1.CreateAlertDictionaryResponse, error) {
	item, err := s.dictionaries.CreateAlertDictionary(ctx, bo.AlertDictionary{Name: req.GetName(), Color: req.GetColor(), Type: req.GetType()})
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertDictionaryResponse{Dictionary: dictionaryItem(*item)}, nil
}

func (s *AlertDictionaryService) UpdateAlertDictionary(ctx context.Context, req *v1.UpdateAlertDictionaryRequest) (*v1.UpdateAlertDictionaryResponse, error) {
	item, err := s.dictionaries.UpdateAlertDictionary(ctx, bo.AlertDictionary{UID: req.GetUid(), Name: req.GetName(), Color: req.GetColor()})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertDictionaryResponse{Dictionary: dictionaryItem(*item)}, nil
}

func (s *AlertDictionaryService) UpdateAlertDictionaryStatus(ctx context.Context, req *v1.UpdateAlertDictionaryStatusRequest) (*v1.UpdateAlertDictionaryStatusResponse, error) {
	item, err := s.dictionaries.UpdateAlertDictionaryStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertDictionaryStatusResponse{Dictionary: dictionaryItem(*item)}, nil
}

func (s *AlertDictionaryService) DeleteAlertDictionary(ctx context.Context, req *v1.DeleteAlertDictionaryRequest) (*v1.DeleteAlertDictionaryResponse, error) {
	if err := s.dictionaries.DeleteAlertDictionary(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertDictionaryResponse{}, nil
}

func (s *AlertDictionaryService) ListAlertDictionaries(ctx context.Context, req *v1.ListAlertDictionariesRequest) (*v1.ListAlertDictionariesResponse, error) {
	items, total, page, pageSize, err := s.dictionaries.ListAlertDictionaries(ctx, req.GetPage(), req.GetPageSize(), req.GetType(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertDictionariesResponse{Items: dictionaryItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}
