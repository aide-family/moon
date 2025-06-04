package bo

import (
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/validate"
)

type TaskScript struct {
	FilePath string        `json:"filePath"`
	FileType vobj.FileType `json:"fileType"`
	Interval time.Duration `json:"interval"`
	Deleted  bool          `json:"deleted"`
}

func (t *TaskScript) IsDeleted() bool {
	if validate.IsNil(t) {
		return true
	}
	return t.Deleted
}
