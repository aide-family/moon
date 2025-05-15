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

func NewFileManagerWithHook(m FileManager, opts ...FileManagerHookOption) FileManager {
	fm := &fileManagerWithHook{
		m:                             m,
		BeforeInitiateMultipartUpload: func(originalName, group string) error { return nil },
		AfterInitiateMultipartUpload:  func(result *InitiateMultipartUploadResult) error { return nil },
		BeforeCompleteMultipartUpload: func(uploadID, objectKey string, parts []UploadPart) error { return nil },
		AfterCompleteMultipartUpload:  func(result *CompleteMultipartUploadResult) error { return nil },
		BeforeGeneratePublicURL:       func(objectKey string) error { return nil },
		AfterGeneratePublicURL:        func(url string) error { return nil },
		BeforeGenerateUploadPartURL:   func(uploadID, objectKey string, partNumber int, expires time.Duration) error { return nil },
		AfterGenerateUploadPartURL:    func(result *UploadPartInfo) error { return nil },
		BeforeDeleteObject:            func(objectKey string) error { return nil },
		AfterDeleteObject:             func(objectKey string) error { return nil },
	}

	for _, opt := range opts {
		opt(fm)
	}

	return fm
}

type FileManagerHookOption func(hook *fileManagerWithHook)

type fileManagerWithHook struct {
	m FileManager

	BeforeInitiateMultipartUpload func(originalName, group string) error
	AfterInitiateMultipartUpload  func(result *InitiateMultipartUploadResult) error

	BeforeCompleteMultipartUpload func(uploadID, objectKey string, parts []UploadPart) error
	AfterCompleteMultipartUpload  func(result *CompleteMultipartUploadResult) error

	BeforeGeneratePublicURL func(objectKey string) error
	AfterGeneratePublicURL  func(url string) error

	BeforeGenerateUploadPartURL func(uploadID, objectKey string, partNumber int, expires time.Duration) error
	AfterGenerateUploadPartURL  func(result *UploadPartInfo) error

	BeforeDeleteObject func(objectKey string) error
	AfterDeleteObject  func(objectKey string) error
}

func (f *fileManagerWithHook) InitiateMultipartUpload(originalName, group string) (result *InitiateMultipartUploadResult, err error) {
	if err = f.BeforeInitiateMultipartUpload(originalName, group); err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = f.AfterInitiateMultipartUpload(result)
		}
	}()
	return f.m.InitiateMultipartUpload(originalName, group)
}

func (f *fileManagerWithHook) GenerateUploadPartURL(uploadID, objectKey string, partNumber int, expires time.Duration) (result *UploadPartInfo, err error) {
	if err := f.BeforeGenerateUploadPartURL(uploadID, objectKey, partNumber, expires); err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = f.AfterGenerateUploadPartURL(result)
		}
	}()
	return f.m.GenerateUploadPartURL(uploadID, objectKey, partNumber, expires)
}

func (f *fileManagerWithHook) CompleteMultipartUpload(uploadID, objectKey string, parts []UploadPart) (result *CompleteMultipartUploadResult, err error) {
	if err := f.BeforeCompleteMultipartUpload(uploadID, objectKey, parts); err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = f.AfterCompleteMultipartUpload(result)
		}
	}()
	return f.m.CompleteMultipartUpload(uploadID, objectKey, parts)
}

func (f *fileManagerWithHook) GeneratePublicURL(objectKey string, exp time.Duration) (result string, err error) {
	if err = f.BeforeGeneratePublicURL(objectKey); err != nil {
		return "", err
	}
	defer func() {
		if err == nil {
			err = f.AfterGeneratePublicURL(result)
		}
	}()
	return f.m.GeneratePublicURL(objectKey, exp)
}

func (f *fileManagerWithHook) DeleteObject(objectKey string) (err error) {
	if err = f.BeforeDeleteObject(objectKey); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = f.AfterDeleteObject(objectKey)
		}
	}()
	return f.m.DeleteObject(objectKey)
}
