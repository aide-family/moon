package repository

import (
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var _ DataRepo = (*UnimplementedDataRepo)(nil)

type (
	DataRepo interface {
		unimplementedDataRepo()
		DB() (*gorm.DB, error)
		Client() (*redis.Client, error)
	}

	UnimplementedDataRepo struct{}
)

func (UnimplementedDataRepo) unimplementedDataRepo() {}

func (UnimplementedDataRepo) DB() (*gorm.DB, error) {
	return nil, status.Error(codes.Unimplemented, "method DB not implemented")
}

func (UnimplementedDataRepo) Client() (*redis.Client, error) {
	return nil, status.Error(codes.Unimplemented, "method Client not implemented")
}
