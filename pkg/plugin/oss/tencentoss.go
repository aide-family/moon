package oss

import (
	"context"
	"io"
	"net/http"
	"net/url"

	tenoss "github.com/tencentyun/cos-go-sdk-v5"
)

// TencentOss  腾讯云对象存储服务
type TencentOss struct {
	client *tenoss.Client
}

// NewTencentOss 创建腾讯云对象存储服务
func NewTencentOss(bucketURL, secretID, secretKey string) (*TencentOss, error) {
	u, _ := url.Parse(bucketURL)
	b := &tenoss.BaseURL{BucketURL: u}
	client := tenoss.NewClient(b, &http.Client{
		Transport: &tenoss.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})
	return &TencentOss{client: client}, nil
}

func (t *TencentOss) UploadFile(objectName string, reader io.Reader, objectSize int64) error {
	_, err := t.client.Object.Put(context.Background(), objectName, reader, nil)
	return err
}

func (t *TencentOss) DownloadFile(objectName string) (io.ReadCloser, error) {
	resp, err := t.client.Object.Get(context.Background(), objectName, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (t *TencentOss) DeleteFile(objectName string) error {
	_, err := t.client.Object.Delete(context.Background(), objectName)
	return err
}
