package file

import (
	"fmt"
	"io"

	"github.com/aide-family/moon/pkg/util/response"
	"github.com/go-kratos/kratos/v2/transport/http"

	fileapi "github.com/aide-family/moon/api/admin/file"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/random"
)

// Service file操作相关服务
type Service struct {
	fileapi.UnimplementedFileManageServer

	fileBiz *biz.FileBiz
}

// NewFileService 创建文件操作服务
func NewFileService(fileBiz *biz.FileBiz) *Service {
	return &Service{
		fileBiz: fileBiz,
	}
}

// UploadFile 上传文件
func (s *Service) UploadFile(ctx http.Context) error {
	file, header, err := ctx.Request().FormFile("file")
	if err != nil {
		return err
	}

	defer file.Close()
	uploadType := ctx.Form().Get("uploadType")
	fileName := fmt.Sprintf("%s-%s", random.UUIDToUpperCase(true), header.Filename)
	params := &bo.ConvertFileParams{
		File:       file,
		UploadType: uploadType,
		Filename:   fileName,
	}
	url, err := s.fileBiz.UploadFile(ctx, params)
	if err != nil {
		return err
	}
	response.Success(ctx, "success", &bo.UploadResParams{URL: url})
	return nil
}

// DownloadFile 下载文件
func (s *Service) DownloadFile(ctx http.Context) error {
	var in bo.DownLoadFileParams
	if err := ctx.BindVars(&in); err != nil {
		return err
	}

	file, err := s.fileBiz.DownLoadFile(ctx, &in)
	if err != nil {
		return err
	}
	defer file.Close()
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", in.FilePath))
	ctx.Response().Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(ctx.Response(), file)
	return err
}
