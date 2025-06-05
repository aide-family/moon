package impl

import (
	"bufio"
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewScriptRepo(bc *conf.Bootstrap) repository.Script {
	return &scriptRepoImpl{
		scriptDirs: bc.GetTaskScripts(),
		scripts:    safety.NewMap[string, *bo.TaskScript](),
	}
}

type scriptRepoImpl struct {
	scriptDirs []string

	scripts *safety.Map[string, *bo.TaskScript]
}

func (s *scriptRepoImpl) GetScripts(ctx context.Context) ([]*bo.TaskScript, error) {
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

func (s *scriptRepoImpl) filterTaskScripts(taskScripts []*bo.TaskScript) []*bo.TaskScript {
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
		oldTaskScript, ok := s.scripts.Get(taskScript.FilePath)
		if !ok {
			filteredTaskScripts = append(filteredTaskScripts, taskScript)
			s.scripts.Set(taskScript.FilePath, taskScript)
			continue
		}
		if oldTaskScript.Hash != taskScript.Hash {
			filteredTaskScripts = append(filteredTaskScripts, taskScript)
			s.scripts.Set(taskScript.FilePath, taskScript)
			continue
		}
	}
	return filteredTaskScripts
}

func (s *scriptRepoImpl) getFiles(dirs ...string) ([]string, error) {
	if len(dirs) == 0 {
		return nil, nil
	}
	fileList := make([]string, 0, len(dirs)*100)
	for _, dir := range dirs {
		var files []string
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		fileList = append(fileList, files...)
	}
	return fileList, nil
}

func (s *scriptRepoImpl) getTaskScripts(files []string) []*bo.TaskScript {
	if len(files) == 0 {
		return nil
	}
	taskScripts := make([]*bo.TaskScript, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		taskScript := &bo.TaskScript{
			FilePath: file,
			FileType: getFileType(file, content),
			Interval: getInterval(file),
		}
		if taskScript.FileType.IsUnknown() || taskScript.Interval < 1*time.Second {
			continue
		}

		taskScript.Content = content
		taskScript.Hash = hash.MD5(string(content))
		taskScripts = append(taskScripts, taskScript)
	}
	return taskScripts
}

func getFileType(file string, content []byte) vobj.FileType {
	switch filepath.Ext(file) {
	case ".sh":
		return vobj.FileTypeShell
	default:
		// Read the first line of the file to determine if it's sh or bash
		reader := bufio.NewReader(bytes.NewReader(content))
		firstLine, err := reader.ReadString('\n')
		if err != nil {
			return vobj.FileTypeUnknown
		}
		if strings.HasPrefix(firstLine, "#!") {
			switch {
			case strings.Contains(firstLine, "bash"):
				return vobj.FileTypeBash
			case strings.Contains(firstLine, "sh"):
				return vobj.FileTypeShell
			case strings.Contains(firstLine, "python"):
				return vobj.FileTypePython
			case strings.Contains(firstLine, "python3"):
				return vobj.FileTypePython3
			}
		}
		return vobj.FileTypeUnknown
	}
}

func getInterval(file string) time.Duration {
	parts := strings.Split(filepath.Base(file), "_")
	if len(parts) < 2 {
		return 0
	}
	// Convert 10s, 5s to time.Duration
	interval, err := time.ParseDuration(parts[0])
	if err != nil {
		return 0
	}
	return interval
}
