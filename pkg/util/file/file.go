package file

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// GetFileExtension 获取文件后缀
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetFileType 获取文件类型
func GetFileType(fileName string) string {
	fileExt := GetFileExtension(fileName)
	return strings.Replace(fileExt, ".", "", 1)
}

// GetFileSizeFromPath 通过文件路径获取文件大小（字节）
func GetFileSizeFromPath(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// GetFileSizeFromFile 通过 multipart.File 获取文件大小（字节）
// 常用于 HTTP 文件上传场景
func GetFileSizeFromFile(file multipart.File) (int64, error) {
	// 将文件指针移到文件末尾以获取文件大小
	fileStat, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = file.Seek(0, io.SeekStart) // 将指针重置回文件开头
	return fileStat, err
}
