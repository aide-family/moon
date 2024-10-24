package oss

import (
	"io"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliOSS struct {
	client     *alioss.Client
	bucketName string
}

// NewAliOSS 创建阿里云OSS客户端
func NewAliOSS(endpoint, accessKeyID, accessKeySecret, bucketName string) (*AliOSS, error) {
	client, err := alioss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	return &AliOSS{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (a *AliOSS) UploadFile(objectName string, reader io.Reader, objectSize int64) error {
	bucket, err := a.client.Bucket(a.bucketName)
	if err != nil {
		return err
	}
	return bucket.PutObject(objectName, reader)
}

func (a *AliOSS) DownloadFile(objectName string) (io.ReadCloser, error) {
	bucket, err := a.client.Bucket(a.bucketName)
	if err != nil {
		return nil, err
	}
	return bucket.GetObject(objectName)
}

func (a *AliOSS) DeleteFile(objectName string) error {
	bucket, err := a.client.Bucket(a.bucketName)
	if err != nil {
		return err
	}
	return bucket.DeleteObject(objectName)
}
