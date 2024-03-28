// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmhistory"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmpage"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmrealtime"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/api"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/captcha"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/chatgroup"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/dashboard"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/dataimpl"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/endpoint"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/msg"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/notify"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/ping"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/promdict"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/role"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/strategy"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/strategygroup"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/syslog"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/user"
	"prometheus-manager/app/prom_server/internal/server"
	"prometheus-manager/app/prom_server/internal/service"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
	"prometheus-manager/app/prom_server/internal/service/authservice"
	"prometheus-manager/app/prom_server/internal/service/dashboardservice"
	"prometheus-manager/app/prom_server/internal/service/interflowservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"
	"prometheus-manager/app/prom_server/internal/service/systemservice"
	"prometheus-manager/pkg/helper/plog"
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
	apiWhite := bootstrap.ApiWhite
	httpServer := server.NewHTTPServer(confServer, dataData, apiWhite, logger)
	pingRepo := ping.NewPingRepo(dataData, logger)
	pingBiz := biz.NewPingRepo(pingRepo, logger)
	pingService := service.NewPingService(pingBiz, logger)
	promDictRepo := promdict.NewPromDictRepo(dataData, logger)
	sysLogRepo := syslog.NewSysLogRepo(dataData, logger)
	dictBiz := biz.NewDictBiz(promDictRepo, sysLogRepo, logger)
	pageRepo := alarmpage.NewAlarmPageRepo(dataData, logger)
	alarmRealtimeRepo := alarmrealtime.NewAlarmRealtime(dataData, logger)
	alarmPageBiz := biz.NewPageBiz(pageRepo, alarmRealtimeRepo, sysLogRepo, logger)
	systemserviceService := systemservice.NewDictService(dictBiz, alarmPageBiz, logger)
	v := data.GetWriteChangeGroupChannel()
	v2 := data.GetWriteRemoveGroupChannel()
	strategyGroupRepo := strategygroup.NewStrategyGroupRepo(dataData, v, v2, logger)
	strategyRepo := strategy.NewStrategyRepo(dataData, v, v2, strategyGroupRepo, logger)
	notifyRepo := notify.NewNotifyRepo(dataData, logger)
	strategyBiz := biz.NewStrategyBiz(strategyRepo, notifyRepo, sysLogRepo, logger)
	msgRepo := msg.NewMsgRepo(dataData, logger)
	userRepo := user.NewUserRepo(dataData, logger)
	notifyBiz := biz.NewNotifyBiz(notifyRepo, sysLogRepo, msgRepo, strategyRepo, userRepo, logger)
	strategyService := promservice.NewStrategyService(strategyBiz, notifyBiz, logger)
	strategyGroupBiz := biz.NewStrategyGroupBiz(strategyGroupRepo, sysLogRepo, logger)
	groupService := promservice.NewGroupService(strategyGroupBiz, logger)
	historyRepo := alarmhistory.NewAlarmHistoryRepo(dataData, logger)
	dataRepo := dataimpl.NewDataRepo(dataData, logger)
	alarmRealtimeBiz := biz.NewAlarmRealtime(dataRepo, alarmRealtimeRepo, pageRepo, logger)
	historyBiz := biz.NewHistoryBiz(historyRepo, pageRepo, msgRepo, strategyRepo, alarmRealtimeBiz, sysLogRepo, logger)
	hookService := alarmservice.NewHookService(historyBiz, logger)
	historyService := alarmservice.NewHistoryService(historyBiz, logger)
	roleRepo := role.NewRoleRepo(dataData, logger)
	userBiz := biz.NewUserBiz(userRepo, dataRepo, roleRepo, sysLogRepo, logger)
	captchaRepo := captcha.NewCaptchaRepo(dataData, logger)
	captchaBiz := biz.NewCaptchaBiz(captchaRepo, logger)
	authService := authservice.NewAuthService(userBiz, captchaBiz, logger)
	userService := systemservice.NewUserService(userBiz, logger)
	apiRepo := api.NewApiRepo(dataData, logger)
	roleBiz := biz.NewRoleBiz(roleRepo, apiRepo, dataRepo, sysLogRepo, logger)
	roleService := systemservice.NewRoleService(roleBiz, logger)
	endpointRepo := endpoint.NewEndpointRepo(dataData, logger)
	endpointBiz := biz.NewEndpointBiz(endpointRepo, sysLogRepo, logger)
	endpointService := promservice.NewEndpointService(endpointBiz, logger)
	apiBiz := biz.NewApiBiz(apiRepo, dataRepo, sysLogRepo, logger)
	apiService := systemservice.NewApiService(apiBiz, logger)
	chatGroupRepo := chatgroup.NewChatGroupRepo(dataData, logger)
	chatGroupBiz := biz.NewChatGroupBiz(chatGroupRepo, sysLogRepo, logger)
	chatGroupService := promservice.NewChatGroupService(chatGroupBiz, notifyBiz, logger)
	notifyService := promservice.NewNotifyService(notifyBiz, logger)
	realtimeService := alarmservice.NewRealtimeService(alarmRealtimeBiz, logger)
	hookInterflowService := interflowservice.NewHookInterflowService(logger)
	dashboardRepo := dashboard.NewDashboardRepo(dataData)
	chartRepo := dashboard.NewChartRepo(dataData)
	dashboardBiz := biz.NewDashboardBiz(dashboardRepo, chartRepo, sysLogRepo, logger)
	chartService := dashboardservice.NewChartService(dashboardBiz, logger)
	dashboardService := dashboardservice.NewDashboardService(dashboardBiz, logger)
	sysLogBiz := biz.NewSysLogBiz(sysLogRepo, logger)
	syslogService := systemservice.NewSyslogService(sysLogBiz, logger)
	notifyTemplateRepo := notify.NewNotifyTemplateRepo(dataData, logger)
	notifyTemplateBiz := biz.NewNotifyTemplateBiz(notifyTemplateRepo, logger)
	templateService := promservice.NewTemplateService(notifyTemplateBiz, logger)
	serverHttpServer := server.RegisterHttpServer(httpServer, pingService, systemserviceService, strategyService, groupService, hookService, historyService, authService, userService, roleService, endpointService, apiService, chatGroupService, notifyService, realtimeService, hookInterflowService, chartService, dashboardService, syslogService, templateService)
	grpcServer := server.NewGRPCServer(confServer, dataData, apiWhite, logger)
	serverGrpcServer := server.RegisterGrpcServer(grpcServer, pingService, systemserviceService, strategyService, groupService, hookService, historyService, userService, roleService, endpointService, apiService, chatGroupService, notifyService, realtimeService, chartService, dashboardService, syslogService)
	v3 := data.GetReadChangeGroupChannel()
	v4 := data.GetReadRemoveGroupChannel()
	alarmEvent, err := server.NewAlarmEvent(dataData, v3, v4, hookService, groupService, logger)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	websocketServer := server.NewWebsocketServer(confServer, logger)
	serverServer := server.NewServer(serverHttpServer, serverGrpcServer, alarmEvent, websocketServer)
	app := newApp(serverServer, logger)
	return app, func() {
		cleanup()
	}, nil
}
