package oss

import (
	"context"
	"io"
)

type (
	// Client oss 客户端
	Client interface {
		// UploadFile 上传文件
		UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error
		// DownloadFile 下载文件
		DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)
		// DeleteFile 删除文件
		DeleteFile(ctx context.Context, objectName string) error
		// GetFileURL 获取文件访问链接
		GetFileURL(ctx context.Context, objectName string) (string, error)
	}
)
