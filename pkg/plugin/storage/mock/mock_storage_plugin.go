//go:build plugin

package main

import (
	"time"

	"github.com/aide-family/moon/pkg/plugin"
	"github.com/aide-family/moon/pkg/plugin/storage"
	"github.com/go-kratos/kratos/v2/log"
)

var _ storage.FileManager = (*MockStorage)(nil)

type MockStorage struct {
	helper *log.Helper
}

func (m *MockStorage) InitiateMultipartUpload(originalName, group string) (*storage.InitiateMultipartUploadResult, error) {
	m.helper.Debugf("InitiateMultipartUpload called with originalName: %s, group: %s", originalName, group)
	return &storage.InitiateMultipartUploadResult{}, nil
}

func (m *MockStorage) GenerateUploadPartURL(uploadID, objectKey string, partNumber int, expires time.Duration) (*storage.UploadPartInfo, error) {
	m.helper.Debugf("GenerateUploadPartURL called with uploadID: %s, objectKey: %s, partNumber: %d, expires: %v", uploadID, objectKey, partNumber, expires)
	return &storage.UploadPartInfo{}, nil
}

func (m *MockStorage) CompleteMultipartUpload(uploadID, objectKey string, parts []storage.UploadPart) (*storage.CompleteMultipartUploadResult, error) {
	m.helper.Debugf("CompleteMultipartUpload called with uploadID: %s, objectKey: %s, parts: %v", uploadID, objectKey, parts)
	return &storage.CompleteMultipartUploadResult{}, nil
}

func (m *MockStorage) GeneratePublicURL(objectKey string, exp time.Duration) (string, error) {
	m.helper.Debugf("GeneratePublicURL called with objectKey: %s, exp: %v", objectKey, exp)
	return "", nil
}

func (m *MockStorage) DeleteObject(objectKey string) error {
	m.helper.Debugf("DeleteObject called with objectKey: %s", objectKey)
	return nil
}

// New is the exported plugin factory function
// Note: This must exactly match the expected signature in the main program
func New(config *plugin.LoadConfig) (storage.FileManager, error) {
	return &MockStorage{
		helper: log.NewHelper(log.With(config.Logger, "module", "plugin.storage.mock")),
	}, nil
}
