package event

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
)

var _ do.Realtime = (*Realtime)(nil)

const tableNameRealtime = "team_realtime_alerts"

type Realtime struct {
	ID           uint32           `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	CreatedAt    time.Time        `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:creation time" json:"createdAt"`
	UpdatedAt    time.Time        `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:update time" json:"updatedAt"`
	TeamID       uint32           `gorm:"column:team_id;type:int;not null;comment:team ID;uniqueIndex:uk__team_id__fingerprint" json:"teamId"`
	Status       vobj.AlertStatus `gorm:"column:status;type:tinyint;not null;comment:status" json:"status"`
	Fingerprint  string           `gorm:"column:fingerprint;type:varchar(255);not null;comment:fingerprint;uniqueIndex:uk__team_id__fingerprint" json:"fingerprint"`
	Labels       kv.StringMap     `gorm:"column:labels;type:text;not null;comment:labels" json:"labels"`
	Summary      string           `gorm:"column:summary;type:text;not null;comment:summary" json:"summary"`
	Description  string           `gorm:"column:description;type:text;not null;comment:description" json:"description"`
	Value        string           `gorm:"column:value;type:text;not null;comment:value" json:"value"`
	GeneratorURL string           `gorm:"column:generator_url;type:text;not null;comment:generator URL" json:"generatorURL"`
	StartsAt     time.Time        `gorm:"column:starts_at;type:datetime;not null;default:'0001-01-01 00:00:00';comment:start time" json:"startsAt"`
	EndsAt       time.Time        `gorm:"column:ends_at;type:datetime;not null;default:'0001-01-01 00:00:00';comment:end time" json:"endsAt"`
}

// GetCreatedAt implements do.Realtime.
func (r *Realtime) GetCreatedAt() time.Time {
	return r.CreatedAt
}

// GetDescription implements do.Realtime.
func (r *Realtime) GetDescription() string {
	return r.Description
}

// GetEndsAt implements do.Realtime.
func (r *Realtime) GetEndsAt() time.Time {
	return r.EndsAt
}

// GetFingerprint implements do.Realtime.
func (r *Realtime) GetFingerprint() string {
	return r.Fingerprint
}

// GetGeneratorURL implements do.Realtime.
func (r *Realtime) GetGeneratorURL() string {
	return r.GeneratorURL
}

// GetID implements do.Realtime.
func (r *Realtime) GetID() uint32 {
	return r.ID
}

// GetLabels implements do.Realtime.
func (r *Realtime) GetLabels() kv.StringMap {
	return r.Labels
}

// GetStartsAt implements do.Realtime.
func (r *Realtime) GetStartsAt() time.Time {
	return r.StartsAt
}

// GetStatus implements do.Realtime.
func (r *Realtime) GetStatus() vobj.AlertStatus {
	return r.Status
}

// GetSummary implements do.Realtime.
func (r *Realtime) GetSummary() string {
	return r.Summary
}

// GetTeamID implements do.Realtime.
func (r *Realtime) GetTeamID() uint32 {
	return r.TeamID
}

// GetUpdatedAt implements do.Realtime.
func (r *Realtime) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

// GetValue implements do.Realtime.
func (r *Realtime) GetValue() string {
	return r.Value
}

func (r *Realtime) TableName() string {
	return genRealtimeTableName(r.TeamID, r.StartsAt)
}

func createRealtimeTable(teamID uint32, t time.Time, tx *gorm.DB) (err error) {
	tableName := genRealtimeTableName(teamID, t)
	if do.HasTable(teamID, tx, tableName) {
		return
	}
	r := &Realtime{
		TeamID:   teamID,
		StartsAt: t,
	}

	if err := do.CreateTable(teamID, tx, tableName, r); err != nil {
		return err
	}
	return
}

func genRealtimeTableName(teamID uint32, t time.Time) string {
	weekStart := do.GetPreviousMonday(t)
	return fmt.Sprintf("%s_%d_%s", tableNameRealtime, teamID, weekStart.Format("20060102"))
}

func GetRealtimeTableName(teamID uint32, t time.Time, tx *gorm.DB) (string, error) {
	tableName := genRealtimeTableName(teamID, t)
	if !do.HasTable(teamID, tx, tableName) {
		return tableName, createRealtimeTable(teamID, t, tx)
	}
	return tableName, nil
}

func GetRealtimeTableNames(teamID uint32, start, end time.Time, tx *gorm.DB) []string {
	// Validate time range
	if start.After(end) {
		return nil
	}

	var tableNames []string

	// Find the first Monday (Monday containing or before start)
	firstMonday := do.GetPreviousMonday(start)

	// From the first Monday, add 7 days each week until exceeding end time
	for currentMonday := firstMonday; !currentMonday.After(end); currentMonday = currentMonday.AddDate(0, 0, 7) {
		// Ensure the generated table name is within the time range (Monday + 6 days not before start)
		if currentMonday.AddDate(0, 0, 6).Before(start) {
			continue
		}
		if do.HasTable(teamID, tx, genRealtimeTableName(teamID, currentMonday)) {
			tableNames = append(tableNames, genRealtimeTableName(teamID, currentMonday))
		}
	}

	return tableNames
}
