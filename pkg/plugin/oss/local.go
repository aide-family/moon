package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aide-family/moon/pkg/conf"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	conf *conf.LocalStorage
}

// NewLocalStorage 创建一个新的本地存储实例
func NewLocalStorage(conf *conf.LocalStorage) *LocalStorage {
	return &LocalStorage{conf: conf}
}

// UploadFile 将文件存储到本地
func (l *LocalStorage) UploadFile(_ context.Context, objectName string, reader io.Reader, _ int64) error {
	fullPath := filepath.Join(l.conf.GetPath(), objectName)

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
func (l *LocalStorage) DownloadFile(_ context.Context, objectName string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.conf.GetPath(), objectName)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// DeleteFile 删除本地文件
func (l *LocalStorage) DeleteFile(_ context.Context, objectName string) error {
	fullPath := filepath.Join(l.conf.GetPath(), objectName)
	return os.Remove(fullPath)
}

// GetFileURL 获取文件的URL
func (l *LocalStorage) GetFileURL(_ context.Context, objectName string) (string, error) {
	fileURL := fmt.Sprintf("%s/%s/%s", l.conf.GetUrl(), l.conf.GetDownloadPre(), objectName)
	return fileURL, nil
}
