package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aide-family/moon/pkg/conf"
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliOSS 阿里云OSS
type AliOSS struct {
	client *alioss.Client
	conf   *conf.AliOss
}

// NewAliOSS 创建阿里云OSS客户端
func NewAliOSS(conf *conf.AliOss) (*AliOSS, error) {
	client, err := alioss.New(conf.GetEndpoint(), conf.GetAccessKeyID(), conf.GetAccessKeySecret())
	if err != nil {
		return nil, err
	}
	return &AliOSS{
		client: client,
		conf:   conf,
	}, nil
}

// UploadFile 上传文件
func (a *AliOSS) UploadFile(_ context.Context, objectName string, reader io.Reader, _ int64) error {
	bucket, err := a.client.Bucket(a.conf.GetBucketName())
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, reader)
}

// DownloadFile 下载文件
func (a *AliOSS) DownloadFile(_ context.Context, objectName string) (io.ReadCloser, error) {
	bucket, err := a.client.Bucket(a.conf.GetBucketName())
	if err != nil {
		return nil, err
	}
	return bucket.GetObject(objectName)
}

// DeleteFile 删除文件
func (a *AliOSS) DeleteFile(_ context.Context, objectName string) error {
	bucket, err := a.client.Bucket(a.conf.GetBucketName())
	if err != nil {
		return err
	}
	return bucket.DeleteObject(objectName)
}

// GetFileURL 获取文件URL
func (a *AliOSS) GetFileURL(_ context.Context, objectName string) (string, error) {
	url := fmt.Sprintf("https://%s.%s/%s", a.conf.GetBucketName(), a.conf.GetEndpoint(), objectName)
	return url, nil
}
