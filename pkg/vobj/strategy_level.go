package vobj

import (
	"database/sql"
	"database/sql/driver"

	"github.com/aide-family/moon/pkg/util/types"
)

var _ sql.Scanner = (*StrategyRawLevel)(nil)
var _ driver.Valuer = (*StrategyRawLevel)(nil)

// StrategyRawLevel strategy level
type StrategyRawLevel struct {
	rawInfo string // rawInfo
}

// NewStrategyLevel new strategy level
func NewStrategyLevel(rawInfo string) *StrategyRawLevel {
	return &StrategyRawLevel{
		rawInfo: rawInfo,
	}
}

// Scan scan value into struct from database driver
func (s *StrategyRawLevel) Scan(src any) (err error) {
	switch v := src.(type) {
	case []byte:
		err = types.Unmarshal(v, &s.rawInfo)
	case string:
		err = types.Unmarshal([]byte(v), &s.rawInfo)
	default:
		err = ErrUnsupportedType
	}
	return err
}

// Value return json value, implement driver.Valuer interface
func (s *StrategyRawLevel) Value() (driver.Value, error) {
	return s.GetRawInfo(), nil
}

// GetRawInfo get raw info
func (s *StrategyRawLevel) GetRawInfo() string {
	if types.IsNil(s) || s.rawInfo == "" {
		return "[]"
	}
	return s.rawInfo
}
