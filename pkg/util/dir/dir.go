package dir

import (
	"os"
	"strings"
)

func MakeDirs(absolutePath string, targetPath ...string) []string {
	configPath := RemoveLastSlash(absolutePath)

	newDirList := make([]string, 0, len(targetPath))
	for _, dir := range targetPath {
		newDir := dir
		// 判断是否为绝对路径
		if !IsAbsolutePath(newDir) {
			newDir = strings.Join([]string{configPath, dir}, "/")
		}

		newDirList = append(newDirList, RemoveLastSlash(newDir))
	}

	return newDirList
}

func MakeDir(absolutePath string, targetPath string) string {
	configPath := absolutePath
	// 去除configPath末尾的"/"
	if configPath != "" && configPath[len(configPath)-1] == '/' {
		configPath = configPath[:len(configPath)-1]
	}

	newDir := targetPath
	// 判断是否为绝对路径
	if !IsAbsolutePath(newDir) {
		newDir = strings.Join([]string{absolutePath, targetPath}, "/")
	}

	// 如果路径最后一个字符是"/"，则去除

	return RemoveLastSlash(newDir)
}

// RemoveLastSlash 去除路径最后一个"/"
func RemoveLastSlash(path string) string {
	if path != "" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return path
}

// IsAbsolutePath 判断是否为绝对路径
func IsAbsolutePath(path string) bool {
	return path != "" && path[0] == '/'
}

// IsDir 判断是否为目录
func IsDir(path string) (bool, error) {
	if path == "" {
		return false, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		// 文件不存在
		if os.IsNotExist(err) {
			// 创建目录
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return false, err
			}
		}
	}

	return fi.IsDir(), err
}

// IsFile 判断是否为文件
func IsFile(path string) (bool, error) {
	if path == "" {
		return false, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		// 文件不存在
		return true, nil
	}

	return !fi.IsDir(), err
}

// BuildYamlFilename 构建yaml文件名
func BuildYamlFilename(filename string) string {
	if filename == "" {
		return ""
	}

	if filename[len(filename)-5:] != ".yaml" {
		filename = filename + ".yaml"
	}

	return filename
}
