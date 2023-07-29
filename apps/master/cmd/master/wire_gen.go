// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
	biz2 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/apps/master/internal/data"
	"prometheus-manager/apps/master/internal/server"
	"prometheus-manager/apps/master/internal/service"
	service2 "prometheus-manager/apps/master/internal/service/prom/v1"
	"prometheus-manager/pkg/conn"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(bootstrap *conf.Bootstrap, logger log.Logger) (*kratos.App, func(), error) {
	env := bootstrap.Env
	confServer := bootstrap.Server
	trace := bootstrap.Trace
	tracerProvider := conn.NewTracerProvider(trace, env)
	confData := bootstrap.Data
	dataData, cleanup, err := data.NewData(confData, logger)
	if err != nil {
		return nil, nil, err
	}
	pingRepo := data.NewPingRepo(dataData, logger)
	pingLogic := biz.NewPingLogic(pingRepo, logger)
	pingService := service.NewPingService(pingLogic, logger)
	crudRepo := data.NewCrudRepo(dataData, logger)
	crudLogic := biz.NewCrudLogic(crudRepo, logger)
	crudService := service.NewCrudService(crudLogic, logger)
	dirRepo := data.NewDirRepo(dataData, logger)
	dirLogic := biz2.NewDirLogic(dirRepo, logger)
	dirService := service2.NewDirService(dirLogic, logger)
	fileRepo := data.NewFileRepo(dataData, logger)
	fileLogic := biz2.NewFileLogic(fileRepo, logger)
	fileService := service2.NewFileService(fileLogic, logger)
	nodeRepo := data.NewNodeRepo(dataData, logger)
	nodeLogic := biz2.NewNodeLogic(nodeRepo, logger)
	nodeService := service2.NewNodeService(nodeLogic, logger)
	ruleRepo := data.NewRuleRepo(dataData, logger)
	ruleLogic := biz2.NewRuleLogic(ruleRepo, logger)
	ruleService := service2.NewRuleService(ruleLogic, logger)
	groupRepo := data.NewGroupRepo(dataData, logger)
	groupLogic := biz2.NewGroupLogic(groupRepo, logger)
	groupService := service2.NewGroupService(groupLogic, logger)
	grpcServer := server.NewGRPCServer(confServer, logger, tracerProvider, pingService, crudService, dirService, fileService, nodeService, ruleService, groupService)
	httpServer := server.NewHTTPServer(confServer, logger, tracerProvider, pingService, crudService, dirService, fileService, nodeService, ruleService, groupService)
	app := newApp(env, logger, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
