package impl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewScriptImpl(bc *conf.Bootstrap) repository.Script {
	return &scriptImpl{
		scriptDirs: bc.GetTaskScripts(),
		scripts:    safety.NewMap[string, *bo.TaskScript](),
	}
}

type scriptImpl struct {
	scriptDirs []string

	scripts *safety.Map[string, *bo.TaskScript]
}

func (s *scriptImpl) GetScripts(ctx context.Context) ([]*bo.TaskScript, error) {
	if len(s.scriptDirs) == 0 {
		return nil, nil
	}
	files, err := s.getFiles(s.scriptDirs...)
	if err != nil {
		return nil, err
	}
	taskScripts := s.getTaskScripts(files)

	taskScripts = s.filterTaskScripts(taskScripts)
	return taskScripts, nil
}

func (s *scriptImpl) filterTaskScripts(taskScripts []*bo.TaskScript) []*bo.TaskScript {
	filteredTaskScripts := make([]*bo.TaskScript, 0, len(taskScripts))

	scripts := s.scripts.List()
	scriptsMap := slices.ToMap(taskScripts, func(v *bo.TaskScript) string {
		return v.FilePath
	})
	for _, taskScript := range scripts {
		if _, ok := scriptsMap[taskScript.FilePath]; !ok {
			taskScript.Deleted = true
			filteredTaskScripts = append(filteredTaskScripts, taskScript)
		}
	}
	for _, taskScript := range taskScripts {
		if _, ok := s.scripts.Get(taskScript.FilePath); !ok {
			filteredTaskScripts = append(filteredTaskScripts, taskScript)
			s.scripts.Set(taskScript.FilePath, taskScript)
		}
	}
	return filteredTaskScripts
}

func (s *scriptImpl) getFiles(dirs ...string) ([]string, error) {
	if len(dirs) == 0 {
		return nil, nil
	}
	fileList := make([]string, 0, len(dirs)*100)
	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir))
		if err != nil {
			return nil, err
		}
		fileList = append(fileList, files...)
	}
	return fileList, nil
}

func (s *scriptImpl) getTaskScripts(files []string) []*bo.TaskScript {
	if len(files) == 0 {
		return nil
	}
	taskScripts := make([]*bo.TaskScript, 0, len(files))
	for _, file := range files {
		taskScript := &bo.TaskScript{
			FilePath: file,
			FileType: getFileType(file),
			Interval: getInterval(file),
		}
		if taskScript.FileType.IsUnknown() || taskScript.Interval < 1*time.Second {
			continue
		}
		taskScripts = append(taskScripts, taskScript)
	}
	return taskScripts
}

func getFileType(file string) vobj.FileType {
	switch filepath.Ext(file) {
	case ".sh":
		return vobj.FileTypeShell
	case ".bash":
		return vobj.FileTypeBash
	case ".py":
		return vobj.FileTypePython
	case ".py3":
		return vobj.FileTypePython3
	case "":
		// Read the first line of the file to determine if it's sh or bash
		firstLine, err := os.ReadFile(file)
		if err != nil {
			return vobj.FileTypeUnknown
		}
		if strings.HasPrefix(string(firstLine), "#!") {
			if strings.Contains(string(firstLine), "bash") {
				return vobj.FileTypeBash
			}
			if strings.Contains(string(firstLine), "sh") {
				return vobj.FileTypeShell
			}
		}
		return vobj.FileTypeUnknown
	default:
		return vobj.FileTypeUnknown
	}
}

func getInterval(file string) time.Duration {
	parts := strings.Split(file, "_")
	if len(parts) < 2 {
		return 0
	}
	// Convert 10s, 5s to time.Duration
	interval, err := time.ParseDuration(parts[1])
	if err != nil {
		return 0
	}
	return interval
}
