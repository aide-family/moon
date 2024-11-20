package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aide-family/moon/pkg/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient MinIO客户端
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	secure     bool
	conf       *conf.Minio
}

// NewMinIO 创建MinIO客户端
func NewMinIO(minioConf *conf.Minio) (*MinIOClient, error) {
	client, err := minio.New(minioConf.GetEndpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(minioConf.GetAccessKeyID(), minioConf.GetAccessKeySecret(), ""),
		Secure: minioConf.GetSecure(), // 根据需要启用 HTTPS
	})
	if err != nil {
		return nil, err
	}

	// init minio bucket
	bucketName := minioConf.BucketName
	// 确保 Bucket 存在，否则创建
	exists, err := client.BucketExists(context.Background(), bucketName)

	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		// 设置 Bucket 为公共访问
		if err = setBucketPublic(context.Background(), client, bucketName); err != nil {
			log.Warn("Error setting bucket public:", err)
			return nil, err
		}
	}

	return &MinIOClient{
		client:     client,
		bucketName: bucketName,
		conf:       minioConf,
	}, nil
}

// UploadFile 上传文件
func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error {
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{})
	return err
}

// DownloadFile 下载文件
func (m *MinIOClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
}

// DeleteFile 删除文件
func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}

// setBucketPublic 设置存储桶为公共访问
func setBucketPublic(ctx context.Context, minioClient *minio.Client, bucketName string) error {
	policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "AllowPublicRead",
                "Effect": "Allow",
                "Principal": {
                    "AWS": ["*"]
                },
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::` + bucketName + `/*"]
            }
        ]
    }`
	// 设置存储桶的访问策略
	return minioClient.SetBucketPolicy(ctx, bucketName, policy)
}

// getFileURL 获取文件的URL
func (m *MinIOClient) getFileURL(objectName string) string {
	secure := "http"
	if m.secure {
		secure = "https"
	}
	fileURL := fmt.Sprintf("%s://%s/%s/%s", secure, m.conf.GetEndpoint(), m.conf.GetBucketName(), objectName)
	return fileURL
}

// GetFileURL 获取文件URL
func (m *MinIOClient) GetFileURL(_ context.Context, objectName string) (string, error) {
	return m.getFileURL(objectName), nil
}
