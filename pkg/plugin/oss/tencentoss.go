package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	tenoss "github.com/tencentyun/cos-go-sdk-v5"
)

// TencentOss  腾讯云对象存储服务
type TencentOss struct {
	client *tenoss.Client
	conf   *conf.TencentOss
}

// NewTencentOss 创建腾讯云对象存储服务
func NewTencentOss(conf *conf.TencentOss) (*TencentOss, error) {
	u, _ := url.Parse(conf.GetBucketURL())
	b := &tenoss.BaseURL{BucketURL: u}
	client := tenoss.NewClient(b, &http.Client{
		Transport: &tenoss.AuthorizationTransport{
			SecretID:  conf.GetSecretID(),
			SecretKey: conf.GetSecretKey(),
		},
	})
	return &TencentOss{conf: conf, client: client}, nil
}

// UploadFile 上传文件
func (t *TencentOss) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error {
	_, err := t.client.Object.Put(ctx, objectName, reader, nil)
	return err
}

// DownloadFile 下载文件
func (t *TencentOss) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	resp, err := t.client.Object.Get(ctx, objectName, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DeleteFile 删除文件
func (t *TencentOss) DeleteFile(ctx context.Context, objectName string) error {
	_, err := t.client.Object.Delete(ctx, objectName)
	return err
}

// GetFileURL 获取文件URL
func (t *TencentOss) GetFileURL(_ context.Context, objectName string) (string, error) {
	conf := t.conf
	return t.GenerateCOSURL(conf.GetBucketName(), conf.GetRegion(), conf.GetSecretID(), conf.GetSecretKey(), objectName, conf.GetIsPublic(), time.Duration(conf.GetExpiry()))
}

// GenerateCOSURL 生成腾讯云 COS 文件的访问 URL
func (t *TencentOss) GenerateCOSURL(bucketName, region, secretID, secretKey, objectKey string, isPublic bool, expiry time.Duration) (string, error) {
	// 初始化 COS 客户端
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, region)
	client := t.client
	if isPublic {
		// 公有存储桶，直接拼接 URL
		return fmt.Sprintf("%s/%s", bucketURL, objectKey), nil
	}
	// 私有存储桶，生成带签名的临时访问 URL
	signedURL, err := client.Object.GetPresignedURL(context.Background(), http.MethodGet, objectKey, secretID, secretKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return signedURL.String(), nil
}
