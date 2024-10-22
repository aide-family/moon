package oss

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient MinIO客户端
type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

// NewMinIO 创建MinIO客户端
func NewMinIO(endpoint, accessKeyID, secretAccessKey, bucketName string, secure bool) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: secure, // 根据需要启用 HTTPS
	})
	if err != nil {
		return nil, err
	}
	return &MinIOClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (m *MinIOClient) UploadFile(objectName string, reader io.Reader, objectSize int64) error {
	_, err := m.client.PutObject(context.Background(), m.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{})
	return err
}

func (m *MinIOClient) DownloadFile(objectName string) (io.ReadCloser, error) {
	return m.client.GetObject(context.Background(), m.bucketName, objectName, minio.GetObjectOptions{})
}

func (m *MinIOClient) DeleteFile(objectName string) error {
	return m.client.RemoveObject(context.Background(), m.bucketName, objectName, minio.RemoveObjectOptions{})
}
