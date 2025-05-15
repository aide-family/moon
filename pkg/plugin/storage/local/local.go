package local

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/plugin/storage"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/go-kratos/kratos/v2/log"
)

var _ storage.FileManager = (*Local)(nil)

func NewLocalOSS(c Config) (*Local, error) {
	if err := os.MkdirAll(c.GetRoot(), 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}
	return &Local{
		c:            c,
		root:         c.GetRoot(),
		uploadMethod: strings.ToUpper(c.GetUploadMethod()),
		uploadURL:    c.GetUploadURL(),
		previewURL:   c.GetPreviewURL(),
		endpoint:     c.GetEndpoint(),
		uploads:      safety.NewMap[string, *uploadSession](),
	}, nil
}

type Config interface {
	GetRoot() string
	GetUploadMethod() string
	GetUploadURL() string
	GetPreviewURL() string
	GetEndpoint() string
}

type Local struct {
	c Config

	root         string
	uploadMethod string
	uploadURL    string
	previewURL   string
	endpoint     string
	uploads      *safety.Map[string, *uploadSession] // uploadID -> session
}

type uploadSession struct {
	objectKey string
	parts     *safety.Map[int, string] // partNumber -> temp file
	createdAt time.Time
}

func (l *Local) GetConfig() Config {
	return l.c
}

func (l *Local) generateObjectKey(originalName, group string) string {
	return fmt.Sprintf("%s/%s/%d_%s", group, timex.Now().Format("2006_01_02"), timex.Now().UnixNano(), originalName)
}

func (l *Local) InitiateMultipartUpload(originalName, group string) (*storage.InitiateMultipartUploadResult, error) {
	objectKey := l.generateObjectKey(originalName, group)
	uploadID := generateUploadID(objectKey)

	l.uploads.Set(uploadID, &uploadSession{
		objectKey: objectKey,
		parts:     safety.NewMap[int, string](),
		createdAt: timex.Now(),
	})

	return &storage.InitiateMultipartUploadResult{
		UploadID:   uploadID,
		BucketName: "Local",
		ObjectKey:  objectKey,
	}, nil
}

func (l *Local) GenerateUploadPartURL(uploadID, objectKey string, partNumber int, expires time.Duration) (*storage.UploadPartInfo, error) {
	urlParams := url.Values{}
	urlParams.Set("uploadID", uploadID)
	urlParams.Set("partNumber", strconv.Itoa(partNumber))
	urlParams.Set("expires", strconv.FormatInt(int64(expires.Seconds()), 10))
	localURL := fmt.Sprintf("%s%s?%s", l.endpoint, l.uploadURL, urlParams.Encode())
	return &storage.UploadPartInfo{
		UploadID:       uploadID,
		BucketName:     "Local",
		ObjectKey:      objectKey,
		PartNumber:     partNumber,
		UploadURL:      localURL,
		ExpirationTime: timex.Now().Add(expires).Unix(),
	}, nil
}

func (l *Local) CompleteMultipartUpload(uploadID, objectKey string, parts []storage.UploadPart) (*storage.CompleteMultipartUploadResult, error) {
	session, exists := l.uploads.Get(uploadID)
	if !exists {
		return nil, fmt.Errorf("upload session not found")
	}
	if objectKey != session.objectKey {
		return nil, fmt.Errorf("object key mismatch")
	}

	finalPath := filepath.Join(l.root, session.objectKey)
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create final file: %w", err)
	}
	defer func(finalFile *os.File) {
		if err := finalFile.Close(); err != nil {
			log.Warnf("failed to close final file: %v", err)
		}
	}(finalFile)

	hashed := md5.New()
	multiWriter := io.MultiWriter(finalFile, hashed)

	f := func(partNumber int, filePath string) error {
		partFile, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open part file: %w", err)
		}
		defer func(partFile *os.File) {
			if err := partFile.Close(); err != nil {
				log.Warnf("failed to close part file: %v", err)
			}
		}(partFile)
		if _, err := partFile.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek part file: %w", err)
		}

		if _, err := io.Copy(multiWriter, partFile); err != nil {
			return fmt.Errorf("failed to merge part %d: %w", partNumber, err)
		}
		return nil
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	for _, part := range parts {
		partFilePath, ok := session.parts.Get(part.PartNumber)
		if !ok {
			return nil, fmt.Errorf("part %d not found", part.PartNumber)
		}

		if err := f(part.PartNumber, partFilePath); err != nil {
			return nil, err
		}
	}

	eTag := hex.EncodeToString(hashed.Sum(nil))

	if err := os.RemoveAll(filepath.Join(l.root, "tmp", uploadID)); err != nil {
		return nil, fmt.Errorf("failed to clean temp files: %w", err)
	}

	l.uploads.Delete(uploadID)
	publicURL, err := l.GeneratePublicURL(session.objectKey, time.Hour*24*7)
	if err != nil {
		return nil, err
	}

	return &storage.CompleteMultipartUploadResult{
		Location:   finalPath,
		Bucket:     "Local",
		Key:        session.objectKey,
		ETag:       eTag,
		PrivateURL: finalPath,
		PublicURL:  publicURL,
	}, nil
}

func (l *Local) GeneratePublicURL(objectKey string, exp time.Duration) (string, error) {
	urlParams := url.Values{}
	urlParams.Set("objectKey", objectKey)
	urlParams.Set("expires", strconv.FormatInt(int64(exp.Seconds()), 10))
	localURL := fmt.Sprintf("%s%s?%s", l.endpoint, l.previewURL, urlParams.Encode())
	return localURL, nil
}

func (l *Local) DeleteObject(objectKey string) error {
	return os.RemoveAll(filepath.Join(l.root, objectKey))
}

func generateUploadID(objectKey string) string {
	h := md5.New()
	h.Write([]byte(objectKey))
	h.Write([]byte(timex.Now().String()))
	return hex.EncodeToString(h.Sum(nil))
}
