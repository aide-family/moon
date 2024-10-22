package file

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/transport/http"
	"io"

	fileapi "github.com/aide-family/moon/api/admin/file"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	httpres "github.com/aide-family/moon/pkg/response"
	"github.com/aide-family/moon/pkg/util/random"
)

// Service file操作相关服务
type Service struct {
	fileapi.UnimplementedFileServer

	fileBiz *biz.FileBiz
}

func NewFileService(fileBiz *biz.FileBiz) *Service {
	return &Service{
		fileBiz: fileBiz,
	}
}

func (s *Service) UploadFile(ctx http.Context) error {
	file, header, err := ctx.Request().FormFile("file")
	if err != nil {
		return err
	}

	defer file.Close()
	uploadType := ctx.Form().Get("uploadType")
	if err != nil {
		return err
	}

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
	httpres.Success(ctx, "success", &bo.UploadResParams{Url: url})
	return nil
}

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
	return nil
}
