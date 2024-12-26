package repoimpl

import (
	"context"
	"io"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/file"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewFileRepository  file repository with the given data source.
func NewFileRepository(data *data.Data) repository.FileRepository {
	return &fileRepositoryImpl{
		data: data,
	}
}

type fileRepositoryImpl struct {
	data *data.Data
}

func (y *fileRepositoryImpl) DownLoadFile(ctx context.Context, params *bo.DownLoadFileParams) (io.ReadCloser, error) {
	return y.data.GetOssCli().DownloadFile(ctx, params.FilePath)
}

func (y *fileRepositoryImpl) UploadFile(ctx context.Context, params *bo.UploadFileParams) (string, error) {
	if !y.data.OssIsOpen() {
		return "", merr.ErrorI18nFileRelatedOssNotOpened(ctx)
	}

	fileName := params.FileName
	fileSize := int64(params.BytesBuff.Len())
	fileSizeMap := y.data.GetFileLimitSize()
	fileExt := file.GetFileType(fileName)

	limitSize, ok := fileSizeMap[fileExt]

	// 未配置限制不允许上传
	if !ok {
		return "", merr.ErrorI18nFileRelatedFileNotSupportedUpload(ctx, fileExt)
	}

	if fileSize > limitSize.GetMax() {
		return "", merr.ErrorI18nFileRelatedFileMaximumLimit(ctx, fileExt)
	}
	err := y.data.GetOssCli().UploadFile(ctx, params.FileName, params.BytesBuff, fileSize)
	if !types.IsNil(err) {
		return "", err
	}
	return y.data.GetOssCli().GetFileURL(ctx, params.FileName)
}
