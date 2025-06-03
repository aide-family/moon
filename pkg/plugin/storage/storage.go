package storage

import (
	"time"
)

type InitiateMultipartUploadResult struct {
	UploadID   string `json:"uploadId"`
	BucketName string `json:"bucketName"`
	ObjectKey  string `json:"objectKey"`
}

type UploadPartInfo struct {
	UploadID       string `json:"uploadId"`
	BucketName     string `json:"bucketName"`
	ObjectKey      string `json:"objectKey"`
	PartNumber     int    `json:"partNumber"`
	UploadURL      string `json:"uploadUrl"`
	ExpirationTime int64  `json:"expirationTime"`
}

type CompleteMultipartUploadResult struct {
	Location   string `json:"location"`
	Bucket     string `json:"bucket"`
	Key        string `json:"key"`
	ETag       string `json:"eTag"`
	PrivateURL string `json:"privateURL"`
	PublicURL  string `json:"publicURL"`
	Expiration int64  `json:"expiration"`
}

type UploadPart struct {
	PartNumber int    `json:"partNumber"` // Part number
	ETag       string `json:"eTag"`       // ETag value of the part's data
}

type FileManager interface {
	InitiateMultipartUpload(originalName, group string) (*InitiateMultipartUploadResult, error)
	GenerateUploadPartURL(uploadID, objectKey string, partNumber int, expires time.Duration) (*UploadPartInfo, error)
	CompleteMultipartUpload(uploadID, objectKey string, parts []UploadPart) (*CompleteMultipartUploadResult, error)
	GeneratePublicURL(objectKey string, exp time.Duration) (string, error)
	DeleteObject(objectKey string) error
}
