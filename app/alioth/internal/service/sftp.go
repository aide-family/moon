package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/sftp"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type SFTPService struct {
	v1.UnimplementedSFTPServiceServer
	sftp *sftp.Usecase
}

func NewSFTPService(uc *sftp.Usecase) *SFTPService {
	return &SFTPService{sftp: uc}
}

func (s *SFTPService) ListSFTP(ctx context.Context, req *v1.ListSFTPRequest) (*v1.ListSFTPResponse, error) {
	entries, audit, err := s.sftp.List(ctx, req.GetAgentUid(), req.GetPath())
	if err != nil {
		return nil, err
	}
	return &v1.ListSFTPResponse{Entries: sftpEntries(entries), Audit: sftpAudit(audit)}, nil
}

func (s *SFTPService) DownloadSFTP(ctx context.Context, req *v1.DownloadSFTPRequest) (*v1.DownloadSFTPResponse, error) {
	content, audit, err := s.sftp.Download(ctx, req.GetAgentUid(), req.GetPath())
	if err != nil {
		return nil, err
	}
	return &v1.DownloadSFTPResponse{Content: content, Audit: sftpAudit(audit)}, nil
}

func (s *SFTPService) UploadSFTP(ctx context.Context, req *v1.UploadSFTPRequest) (*v1.UploadSFTPResponse, error) {
	audit, err := s.sftp.Upload(ctx, req.GetAgentUid(), req.GetPath(), req.GetContent())
	if err != nil {
		return nil, err
	}
	return &v1.UploadSFTPResponse{Audit: sftpAudit(audit)}, nil
}

func (s *SFTPService) MkdirSFTP(ctx context.Context, req *v1.MkdirSFTPRequest) (*v1.MkdirSFTPResponse, error) {
	audit, err := s.sftp.Mkdir(ctx, req.GetAgentUid(), req.GetPath())
	if err != nil {
		return nil, err
	}
	return &v1.MkdirSFTPResponse{Audit: sftpAudit(audit)}, nil
}

func (s *SFTPService) DeleteSFTP(ctx context.Context, req *v1.DeleteSFTPRequest) (*v1.DeleteSFTPResponse, error) {
	audit, err := s.sftp.Delete(ctx, req.GetAgentUid(), req.GetPath())
	if err != nil {
		return nil, err
	}
	return &v1.DeleteSFTPResponse{Audit: sftpAudit(audit)}, nil
}

func (s *SFTPService) ListSFTPAudits(ctx context.Context, req *v1.ListSFTPAuditsRequest) (*v1.ListSFTPAuditsResponse, error) {
	items, total, page, pageSize, err := s.sftp.ListAudits(ctx, req.GetAgentUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.SFTPAuditItem, 0, len(items))
	for i := range items {
		out = append(out, sftpAudit(&items[i]))
	}
	return &v1.ListSFTPAuditsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func sftpEntries(items []bo.SFTPEntry) []*v1.SFTPEntry {
	out := make([]*v1.SFTPEntry, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.SFTPEntry{
			Name: item.Name, Path: item.Path, IsDir: item.IsDir, Size: item.Size, Mode: item.Mode, ModTime: item.ModTime,
		})
	}
	return out
}

func sftpAudit(item *bo.SFTPAudit) *v1.SFTPAuditItem {
	if item == nil {
		return nil
	}
	created := ""
	if !item.CreatedAt.IsZero() {
		created = item.CreatedAt.UTC().Format(time.RFC3339)
	}
	return &v1.SFTPAuditItem{
		Uid: item.UID, AgentUid: item.AgentUID, OperatorUid: item.OperatorUID, Path: item.Path,
		Action: item.Action, Result: item.Result, Detail: item.Detail, CreatedAt: created,
	}
}
