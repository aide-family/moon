package oss

import "io"

type OssClient interface {
	// UploadFile 上传文件
	UploadFile(objectName string, reader io.Reader, objectSize int64) error

	// DownloadFile 下载文件
	DownloadFile(objectName string) (io.ReadCloser, error)
	// DeleteFile 删除文件
	DeleteFile(objectName string) error
}
