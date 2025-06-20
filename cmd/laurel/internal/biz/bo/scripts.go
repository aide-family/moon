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
	Content  []byte        `json:"content"`
	Hash     string        `json:"hash"`
	Deleted  bool          `json:"deleted"`
}

func (t *TaskScript) IsDeleted() bool {
	if validate.IsNil(t) {
		return true
	}
	return t.Deleted
}

// IsValid returns true if the script is valid
func (t *TaskScript) IsValid() bool {
	if validate.IsNil(t) {
		return false
	}
	return !t.FileType.IsUnknown() && t.Interval > 0
}
