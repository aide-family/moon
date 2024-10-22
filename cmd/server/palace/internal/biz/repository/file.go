package repository

import (
	"context"
	"io"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
)

// FileRepository file repository interface
type FileRepository interface {
	// UploadFile 上传文件
	UploadFile(ctx context.Context, params *bo.UploadFileParams) (string, error)

	DownLoadFile(ctx context.Context, params *bo.DownLoadFileParams) (io.ReadCloser, error)
}
