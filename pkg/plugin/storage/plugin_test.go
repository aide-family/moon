package storage_test

import (
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/pkg/plugin"
	"github.com/moon-monitor/moon/pkg/plugin/storage"
)

const mockStoragePlugin = "./mock/mock_storage_plugin.so"

func TestLoadPlugin(t *testing.T) {
	logger := log.NewStdLogger(os.Stdout)

	manager, err := storage.LoadPlugin(&plugin.LoadConfig{
		Path:    mockStoragePlugin,
		Logger:  logger,
		Configs: nil,
	})
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	_, err = manager.InitiateMultipartUpload("test", "test")
	if err != nil {
		t.Errorf("Failed to initiate multipart upload: %v", err)
	}
	_, err = manager.GenerateUploadPartURL("test", "test", 1, 0)
	if err != nil {
		t.Errorf("Failed to generate upload part URL: %v", err)
	}
	_, err = manager.CompleteMultipartUpload("test", "test", []storage.UploadPart{})
	if err != nil {
		t.Errorf("Failed to complete multipart upload: %v", err)
	}
	_, err = manager.GeneratePublicURL("test", 0)
	if err != nil {
		t.Errorf("Failed to generate public URL: %v", err)
	}
	err = manager.DeleteObject("test")
	if err != nil {
		t.Errorf("Failed to delete object: %v", err)
	}
}
