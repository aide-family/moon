package bo

import (
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
)

type TaskScript struct {
	FilePath string        `json:"filePath"`
	FileType vobj.FileType `json:"fileType"`
	Interval time.Duration `json:"interval"`
	Deleted  bool          `json:"deleted"`
}
