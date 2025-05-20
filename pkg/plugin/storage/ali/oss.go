package ali

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/aide-family/moon/pkg/plugin/storage"
	"github.com/aide-family/moon/pkg/util/timex"
)

var _ storage.FileManager = (*aliCloud)(nil)

func NewOSS(c Config) (storage.FileManager, error) {
	client, err := alioss.New(c.GetEndpoint(), c.GetAccessKeyId(), c.GetAccessKeySecret())
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(c.GetBucketName())
	if err != nil {
		return nil, err
	}

	return &aliCloud{
		accessKeyID:     c.GetAccessKeyId(),
		accessKeySecret: c.GetAccessKeySecret(),
		bucketName:      c.GetBucketName(),
		endpoint:        c.GetEndpoint(),
		client:          client,
		bucket:          bucket,
	}, nil
}

type Config interface {
	GetEndpoint() string
	GetAccessKeyId() string
	GetAccessKeySecret() string
	GetBucketName() string
}

type aliCloud struct {
	endpoint        string
	accessKeyID     string
	accessKeySecret string
	bucketName      string

	client *alioss.Client
	bucket *alioss.Bucket
}

func generateObjectKey(originalName, group string) string {
	return fmt.Sprintf("uploads/%s/%s/%d_%s", group, timex.Now().Format("2006_01_02"), timex.Now().UnixNano(), originalName)
}

func (a *aliCloud) InitiateMultipartUpload(originalName, group string) (*storage.InitiateMultipartUploadResult, error) {
	uniqueKey := generateObjectKey(originalName, group)

	// Initialize multipart upload
	multipartUpload, err := a.bucket.InitiateMultipartUpload(uniqueKey)
	if err != nil {
		return nil, err
	}

	return &storage.InitiateMultipartUploadResult{
		UploadID:   multipartUpload.UploadID,
		BucketName: multipartUpload.Bucket,
		ObjectKey:  multipartUpload.Key,
	}, nil
}

func (a *aliCloud) GenerateUploadPartURL(uploadID, objectKey string, partNumber int, expires time.Duration) (*storage.UploadPartInfo, error) {
	sec := int64(expires.Seconds())

	options := []alioss.Option{
		alioss.AddParam("uploadId", uploadID),
		alioss.AddParam("partNumber", strconv.Itoa(partNumber)),
		alioss.ACReqMethod("PUT"),
		alioss.Expires(timex.Now().Add(time.Duration(sec) * time.Second)),
		alioss.ContentType("application/octet-stream"),
	}

	signedURL, err := a.bucket.SignURL(objectKey, alioss.HTTPPut, sec, options...)
	if err != nil {
		return nil, err
	}

	return &storage.UploadPartInfo{
		UploadID:       uploadID,
		BucketName:     a.bucketName,
		ObjectKey:      objectKey,
		PartNumber:     partNumber,
		UploadURL:      signedURL,
		ExpirationTime: timex.Now().Add(time.Duration(sec) * time.Second).Unix(),
	}, nil
}

func (a *aliCloud) CompleteMultipartUpload(uploadID, objectKey string, parts []storage.UploadPart) (*storage.CompleteMultipartUploadResult, error) {
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	multipartUploadResult := alioss.InitiateMultipartUploadResult{
		Bucket:   a.bucketName,
		Key:      objectKey,
		UploadID: uploadID,
	}

	uploadParts := make([]alioss.UploadPart, len(parts))
	for i, part := range parts {
		uploadParts[i] = alioss.UploadPart{
			PartNumber: part.PartNumber,
			ETag:       strings.Trim(part.ETag, "\""),
		}
	}

	completeMultipartUpload, err := a.bucket.CompleteMultipartUpload(multipartUploadResult, uploadParts)
	if err != nil {
		return nil, fmt.Errorf("complete multipart upload failed: %w", err)
	}

	url := fmt.Sprintf("https://%s.%s/%s", a.bucketName, a.endpoint, objectKey)
	publicURL, err := a.GeneratePublicURL(objectKey, time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("generate public URL failed: %w", err)
	}

	return &storage.CompleteMultipartUploadResult{
		Location:   completeMultipartUpload.Location,
		Bucket:     completeMultipartUpload.Bucket,
		Key:        completeMultipartUpload.Key,
		ETag:       completeMultipartUpload.ETag,
		PrivateURL: url,
		PublicURL:  publicURL,
		Expiration: 0,
	}, nil
}

func (a *aliCloud) GeneratePublicURL(objectKey string, exp time.Duration) (string, error) {
	options := []alioss.Option{
		alioss.ResponseContentDisposition(fmt.Sprintf("attachment; filename=\"%s\"", objectKey)),
	}
	expires := timex.Now().Add(exp)
	signedURL, err := a.bucket.SignURL(objectKey, alioss.HTTPGet, int64(expires.Second()), options...)
	if err != nil {
		return "", err
	}

	return signedURL, nil
}

func (a *aliCloud) DeleteObject(objectKey string) error {
	return a.bucket.DeleteObject(objectKey)
}
