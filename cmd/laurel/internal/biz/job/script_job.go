package job

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/command"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/hash"
)

var _ cron_server.CronJob = (*scriptJob)(nil)

func NewScriptJob(script *bo.TaskScript, logger log.Logger) cron_server.CronJob {
	return &scriptJob{
		script: script,
		helper: log.NewHelper(log.With(logger, "module", "job.script", "script", script.FilePath)),
	}
}

type scriptJob struct {
	script *bo.TaskScript
	helper *log.Helper
	id     cron.EntryID
}

// Index implements cron_server.CronJob.
func (s *scriptJob) Index() string {
	return hash.MD5(s.script.FilePath)
}

// IsImmediate implements cron_server.CronJob.
func (s *scriptJob) IsImmediate() bool {
	return true
}

// Spec implements cron_server.CronJob.
func (s *scriptJob) Spec() cron_server.CronSpec {
	if s.script.Interval <= 1*time.Second {
		return cron_server.CronSpecEvery(1 * time.Minute)
	}
	return cron_server.CronSpecEvery(s.script.Interval)
}

// WithID implements cron_server.CronJob.
func (s *scriptJob) WithID(id cron.EntryID) cron_server.CronJob {
	s.id = id
	return s
}

func (s *scriptJob) Run() {
	s.helper.Infof("script job run: %s", s.script.FilePath)
	if s.script.IsDeleted() {
		s.helper.Infof("script job run: %s, deleted", s.script.FilePath)
		return
	}
	var (
		content string
		err     error
	)
	switch s.script.FileType {
	case vobj.FileTypePython:
		content, err = command.ExecPython(string(s.script.Content))
	case vobj.FileTypePython3:
		content, err = command.ExecPython3(string(s.script.Content))
	case vobj.FileTypeShell:
		content, err = command.ExecShell(string(s.script.Content))
	case vobj.FileTypeBash:
		content, err = command.ExecBash(string(s.script.Content))
	default:
		s.helper.Warnf("script job run: %s, file type: %s", s.script.FilePath, s.script.FileType)
		return
	}
	if err != nil {
		s.helper.Errorf("script job run: %s, error: %v", s.script.FilePath, err)
		return
	}
	s.helper.Info(content)
}

func (s *scriptJob) ID() cron.EntryID {
	return s.id
}
