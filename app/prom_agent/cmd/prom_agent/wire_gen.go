// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/aide-family/moon/app/prom_agent/internal/biz"
	"github.com/aide-family/moon/app/prom_agent/internal/conf"
	"github.com/aide-family/moon/app/prom_agent/internal/data"
	"github.com/aide-family/moon/app/prom_agent/internal/data/repositiryimpl"
	"github.com/aide-family/moon/app/prom_agent/internal/server"
	"github.com/aide-family/moon/app/prom_agent/internal/service"
	"github.com/aide-family/moon/pkg/helper/plog"
	"github.com/go-kratos/kratos/v2"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(string2 *string) (*kratos.App, func(), error) {
	confBefore := before()
	bootstrap, err := conf.LoadConfig(string2, confBefore)
	if err != nil {
		return nil, nil, err
	}
	confServer := bootstrap.Server
	log := bootstrap.Log
	logger := plog.NewLogger(log)
	dataData, cleanup, err := data.NewData(bootstrap, logger)
	if err != nil {
		return nil, nil, err
	}
	pingRepo := data.NewPingRepo(dataData, logger)
	pingBiz := biz.NewPingBiz(pingRepo, logger)
	pingService := service.NewPingService(pingBiz, logger)
	grpcServer := server.NewGRPCServer(confServer, pingService, logger)
	hookService := service.NewHookService(logger)
	interflow := bootstrap.Interflow
	hookInterflowService := service.NewHookInterflowService(interflow, logger)
	httpServer := server.NewHTTPServer(confServer, pingService, hookService, hookInterflowService, logger)
	watchProm := bootstrap.WatchProm
	alarmRepo := repositiryimpl.NewAlarmRepo(dataData, interflow, logger)
	alarmBiz := biz.NewAlarmBiz(alarmRepo, logger)
	loadService := service.NewLoadService(alarmBiz, logger)
	watch, err := server.NewWatch(watchProm, dataData, loadService, logger)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	app := newApp(grpcServer, httpServer, watch, logger)
	return app, func() {
		cleanup()
	}, nil
}
