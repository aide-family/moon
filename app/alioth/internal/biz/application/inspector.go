package application

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
)

// Inspector 聚合部署组状态并读取操作记录。
type Inspector struct {
	releases repository.ApplicationReleaseRepo
	log      *slog.Logger
}

func NewInspector(releases repository.ApplicationReleaseRepo, logger *slog.Logger) *Inspector {
	return &Inspector{releases: releases, log: logger}
}

// Recompute 按集群实例状态重新计算部署组状态。
func (inspector *Inspector) Recompute(ctx context.Context, namespaceUID, deploymentUID string) {
	deployment, err := inspector.releases.GetDeployment(ctx, namespaceUID, deploymentUID)
	if err != nil {
		inspector.log.Error("recompute deployment: " + err.Error())
		return
	}
	if len(deployment.Releases) == 0 {
		return
	}
	counts := map[string]int{}
	for _, release := range deployment.Releases {
		counts[release.Status]++
	}
	total := len(deployment.Releases)
	status := "RUNNING"
	messages := make([]string, 0)
	switch {
	case counts["UNINSTALLING"] > 0:
		status = "UNINSTALLING"
	case counts["UNINSTALLED"] == total:
		status = "UNINSTALLED"
	case counts["PENDING"] > 0 || counts["DEPLOYING"] > 0:
		status = "DEPLOYING"
	case counts["RUNNING"] == total:
		status = "RUNNING"
	case counts["FAILED"] == total:
		status = "FAILED"
	default:
		status = "PARTIAL"
	}
	for _, release := range deployment.Releases {
		if release.Status == "FAILED" && release.Message != "" {
			messages = append(messages, release.ClusterName+": "+release.Message)
		}
	}
	deployment.Status = status
	deployment.Message = strings.Join(messages, "; ")
	if err := inspector.releases.UpdateDeployment(ctx, *deployment); err != nil {
		inspector.log.Error("update deployment status: " + err.Error())
	}
}

func (inspector *Inspector) ListOperations(ctx context.Context, namespaceUID, deploymentUID, releaseUID string, pageNo, pageSize int32) ([]bo.ApplicationOperation, int, int, int, error) {
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := inspector.releases.ListOperations(ctx, namespaceUID, deploymentUID, releaseUID, p, s)
	return items, total, p, s, err
}
