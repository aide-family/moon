package ali_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/aide-family/moon/pkg/plugin/storage"
	"github.com/aide-family/moon/pkg/plugin/storage/ali"
)

var _ ali.Config = (*config)(nil)

type config struct {
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	bucketName      string
}

func (c *config) GetEndpoint() string {
	return c.endpoint
}

func (c *config) GetAccessKeyId() string {
	return c.accessKeyID
}

func (c *config) GetAccessKeySecret() string {
	return c.accessKeySecret
}

func (c *config) GetBucketName() string {
	return c.bucketName
}

func Test_NewOSS(t *testing.T) {
	_ = godotenv.Load(".env")
	c := &config{
		endpoint:        os.Getenv("ALIYUN_OSS_ENDPOINT"),
		accessKeyID:     os.Getenv("ALIYUN_OSS_ACCESS_KEY_ID"),
		accessKeySecret: os.Getenv("ALIYUN_OSS_ACCESS_KEY_SECRET"),
		bucketName:      os.Getenv("ALIYUN_OSS_BUCKET_NAME"),
	}
	if c.GetEndpoint() == "" {
		return
	}

	fileName := "test.txt"
	// Create a test.txt file with 2M data
	if err := os.WriteFile(fileName, bytes.Repeat([]byte("a"), 2*1024*1024), 0644); err != nil {
		t.Fatal(err)
	}

	fileManager, err := ali.NewOSS(c)
	if err != nil {
		t.Fatal(err)
	}

	initiateMultipartUpload, err := fileManager.InitiateMultipartUpload(fileName, "test_0")
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
}

//const chunkSize = 5 * 1024 * 1024 // 5MB

const chunkSize = 1 * 1024 * 1024

func uploadFile(fileName string, manager storage.FileManager, params *storage.InitiateMultipartUploadResult) ([]storage.UploadPart, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

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
	defer func() {
		_ = resp.Body.Close()
	}()

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
