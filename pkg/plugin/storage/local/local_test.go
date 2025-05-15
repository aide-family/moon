package local_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/moon-monitor/moon/pkg/plugin/storage"
	"github.com/moon-monitor/moon/pkg/plugin/storage/local"
)

var _ local.Config = (*config)(nil)

type config struct {
	root         string
	uploadMethod string
	uploadURL    string
	previewURL   string
	endpoint     string
}

func (c *config) GetRoot() string {
	return c.root
}

func (c *config) GetUploadMethod() string {
	return c.uploadMethod
}

func (c *config) GetUploadURL() string {
	return c.uploadURL
}

func (c *config) GetPreviewURL() string {
	return c.previewURL
}

func (c *config) GetEndpoint() string {
	return c.endpoint
}

func Test_NewLocalOSS(t *testing.T) {
	c := &config{
		root:         "./moon",
		uploadMethod: "PUT",
		uploadURL:    "/upload",
		previewURL:   "/preview",
		endpoint:     "http://localhost:8080",
	}

	localOSS, err := local.NewLocalOSS(c)
	if err != nil {
		t.Fatal(err)
	}

	opts := []kratoshttp.ServerOption{
		kratoshttp.Address(":8080"),
	}

	srv := kratoshttp.NewServer(opts...)

	route := srv.Route("/")
	route.Handle(c.GetUploadMethod(), c.GetUploadURL(), func(c kratoshttp.Context) error {
		return localOSS.UploadHandler(c.Response(), c.Request())
	})

	route.GET(c.GetPreviewURL(), func(c kratoshttp.Context) error {
		return localOSS.PreviewHandler(c.Response(), c.Request())
	})

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go srv.Start(context.Background())

	go func() {
		defer func() {
			time.Sleep(10 * time.Second)
			ch <- syscall.SIGINT
		}()
		var fileManager storage.FileManager = localOSS
		fileName := "test.txt"
		// Create a test.txt file with 2M data
		if err := os.WriteFile(fileName, bytes.Repeat([]byte("a"), 2*1024*1024), 0644); err != nil {
			t.Fatal(err)
		}

		initiateMultipartUpload, err := fileManager.InitiateMultipartUpload("test.txt", "test_0")
		if err != nil {
			t.Fatal(err)
		}
		parts, err := uploadFile(fileName, fileManager, initiateMultipartUpload)
		if err != nil {
			t.Fatal(err)
		}
		completeMultipartUpload, err := fileManager.CompleteMultipartUpload(initiateMultipartUpload.UploadID, initiateMultipartUpload.ObjectKey, parts)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("Upload completed, visit the address: %s\n", completeMultipartUpload.PublicURL)
	}()

	for {
		select {
		case <-ch:
			return
		}
	}
}

const chunkSize = 1 * 1024 * 1024

func uploadFile(fileName string, manager storage.FileManager, params *storage.InitiateMultipartUploadResult) ([]storage.UploadPart, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file information: %v", err)
	}

	fileSize := fileInfo.Size()
	chunks := int((fileSize + chunkSize - 1) / chunkSize)
	parts := make([]storage.UploadPart, 0, chunks)

	for i := 0; i < chunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize
		if end > fileSize {
			end = fileSize
		}

		fmt.Printf("uploading the %d/%d shard...\n", i+1, chunks)

		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("file location failure: %v", err)
		}

		chunk := make([]byte, end-start)
		_, err = io.ReadFull(file, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to read the file shard: %v", err)
		}

		partNumber := i + 1
		partURL, err := manager.GenerateUploadPartURL(params.UploadID, params.ObjectKey, partNumber, 60*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to get the pre-signed URL: %v", err)
		}

		eTag, err := uploadChunk(partURL.UploadURL, chunk)
		if err != nil {
			return nil, fmt.Errorf("the %d shard upload failed: %v", partNumber, err)
		}

		parts = append(parts, storage.UploadPart{
			PartNumber: partNumber,
			ETag:       eTag,
		})
	}

	return parts, nil
}

func uploadChunk(url string, chunk []byte) (string, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(chunk))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", strconv.Itoa(len(chunk)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to upload shard. HTTP status code: %d", resp.StatusCode)
	}

	eTag := resp.Header.Get("ETag")
	if eTag == "" {
		eTag = resp.Header.Get("etag")
	}
	if eTag == "" {
		eTag = resp.Header.Get("Etag")
	}
	if eTag == "" {
		return "", fmt.Errorf("unable to get ETag from response header")
	}

	return strings.Trim(eTag, "\""), nil
}
