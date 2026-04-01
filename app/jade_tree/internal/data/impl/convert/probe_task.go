// Package convert provides DO/BO conversion helpers for impl repositories.
package convert

import (
	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/data/impl/do"
)

// ToProbeTaskItemBo maps a persisted probe task row to BO.
func ToProbeTaskItemBo(row *do.ProbeTask) *bo.ProbeTaskItemBo {
	if row == nil {
		return nil
	}
	return &bo.ProbeTaskItemBo{
		UID:            row.ID,
		Type:           row.Type,
		Host:           row.Host,
		Port:           row.Port,
		URL:            row.URL,
		Name:           row.Name,
		Status:         row.Status,
		TimeoutSeconds: row.TimeoutSeconds,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
