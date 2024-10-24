package oss

import (
	"io"
	"os"
	"path/filepath"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建一个新的本地存储实例
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// UploadFile 将文件存储到本地
func (l *LocalStorage) UploadFile(objectName string, reader io.Reader, objectSize int64) error {
	fullPath := filepath.Join(l.basePath, objectName)

	// 创建文件所在的目录
	err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	if err != nil {
		return err
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将数据写入文件
	_, err = io.Copy(file, reader)
	return err
}

// DownloadFile 从本地存储中读取文件
func (l *LocalStorage) DownloadFile(objectName string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, objectName)
	return os.Open(fullPath)
}

// DeleteFile 删除本地文件
func (l *LocalStorage) DeleteFile(objectName string) error {
	fullPath := filepath.Join(l.basePath, objectName)
	return os.Remove(fullPath)
}
