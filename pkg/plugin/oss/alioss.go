package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aide-family/moon/pkg/conf"
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

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

func (a *AliOSS) UploadFile(_ context.Context, objectName string, reader io.Reader, _ int64) error {
	bucket, err := a.client.Bucket(a.conf.GetBucketName())
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, reader)
}

func (a *AliOSS) DownloadFile(_ context.Context, objectName string) (io.ReadCloser, error) {
	bucket, err := a.client.Bucket(a.conf.GetBucketName())
	if err != nil {
		return nil, err
	}
	return bucket.GetObject(objectName)
}

func (a *AliOSS) DeleteFile(_ context.Context, objectName string) error {
	bucket, err := a.client.Bucket(a.conf.GetBucketName())
	if err != nil {
		return err
	}
	return bucket.DeleteObject(objectName)
}

func (a *AliOSS) GetFileUrl(_ context.Context, objectName string) (string, error) {
	url := fmt.Sprintf("https://%s.%s/%s", a.conf.GetBucketName(), a.conf.GetEndpoint(), objectName)
	return url, nil
}
