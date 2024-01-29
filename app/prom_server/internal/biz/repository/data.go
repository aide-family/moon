package repository

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"prometheus-manager/pkg/util/cache"
)

var _ DataRepo = (*UnimplementedDataRepo)(nil)

type (
	DataRepo interface {
		unimplementedDataRepo()
		DB() (*gorm.DB, error)
		Cache() (cache.GlobalCache, error)
	}

	UnimplementedDataRepo struct{}
)

func (UnimplementedDataRepo) unimplementedDataRepo() {}

func (UnimplementedDataRepo) DB() (*gorm.DB, error) {
	return nil, status.Error(codes.Unimplemented, "method DB not implemented")
}

func (UnimplementedDataRepo) Cache() (cache.GlobalCache, error) {
	return nil, status.Error(codes.Unimplemented, "method Client not implemented")
}
