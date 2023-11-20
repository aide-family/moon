package strategy

import (
	"encoding/json"
	"testing"

	"github.com/go-kratos/kratos/v2/config/file"
)

func TestLoad_Load(t *testing.T) {
	load := NewStrategyLoad(file.NewSource("."))
	groups, err := load.Load()
	if err != nil {
		t.Error(err)
		return
	}

	groupsByte, _ := json.Marshal(groups)
	t.Log(string(groupsByte))
}
