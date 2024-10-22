package biz

import (
	"bytes"
	"context"
	"io"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	fileUil "github.com/aide-family/moon/pkg/util/file"
)

// NewFileBiz new file biz
func NewFileBiz(fileRepo repository.FileRepository) *FileBiz {
	return &FileBiz{
		fileRepo: fileRepo,
	}
}

// FileBiz file 业务
type FileBiz struct {
	fileRepo repository.FileRepository
}

// UploadFile 解析上传文件
func (y *FileBiz) UploadFile(ctx context.Context, param *bo.ConvertFileParams) (string, error) {
	fileName := param.Filename
	file := param.File

	fileData := &bytes.Buffer{}
	_, err := io.Copy(fileData, file)
	if err != nil {
		return "", err
	}
	uploadParam := &bo.UploadFileParams{
		FileName:  fileName,
		BytesBuff: fileData,
		Extension: fileUil.GetFileExtension(fileName),
	}
	return y.fileRepo.UploadFile(ctx, uploadParam)
}

// DownLoadFile 下载本地文件
func (y *FileBiz) DownLoadFile(ctx context.Context, param *bo.DownLoadFileParams) (io.ReadCloser, error) {
	return y.fileRepo.DownLoadFile(ctx, param)
}
