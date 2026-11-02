package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/board"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// BoardService 把大盘接口转成 biz 入参；布局与权限判断都在 biz。
type BoardService struct {
	v1.UnimplementedBoardServiceServer
	gateway *board.Gateway
}

func NewBoardService(gateway *board.Gateway) *BoardService {
	return &BoardService{gateway: gateway}
}

func (s *BoardService) CreateBoard(ctx context.Context, req *v1.CreateBoardRequest) (*v1.CreateBoardResponse, error) {
	item, err := s.gateway.CreateBoard(ctx, bo.Board{
		Name: req.GetName(), Description: req.GetDescription(), Panels: boardPanelInputs(req.GetPanels()),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateBoardResponse{Board: boardItem(*item)}, nil
}

func (s *BoardService) UpdateBoard(ctx context.Context, req *v1.UpdateBoardRequest) (*v1.UpdateBoardResponse, error) {
	item, err := s.gateway.UpdateBoard(ctx, bo.Board{
		UID: req.GetUid(), Name: req.GetName(), Description: req.GetDescription(), Panels: boardPanelInputs(req.GetPanels()),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateBoardResponse{Board: boardItem(*item)}, nil
}

func (s *BoardService) UpdateBoardStatus(ctx context.Context, req *v1.UpdateBoardStatusRequest) (*v1.UpdateBoardStatusResponse, error) {
	item, err := s.gateway.UpdateBoardStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateBoardStatusResponse{Board: boardItem(*item)}, nil
}

func (s *BoardService) DeleteBoard(ctx context.Context, req *v1.DeleteBoardRequest) (*v1.DeleteBoardResponse, error) {
	if err := s.gateway.DeleteBoard(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteBoardResponse{}, nil
}

func (s *BoardService) GetBoard(ctx context.Context, req *v1.GetBoardRequest) (*v1.GetBoardResponse, error) {
	item, err := s.gateway.GetBoard(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetBoardResponse{Board: boardItem(*item)}, nil
}

func (s *BoardService) ListBoards(ctx context.Context, req *v1.ListBoardsRequest) (*v1.ListBoardsResponse, error) {
	items, total, page, pageSize, err := s.gateway.ListBoards(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.BoardListItem, 0, len(items))
	for i := range items {
		out = append(out, boardListItem(items[i]))
	}
	return &v1.ListBoardsResponse{
		Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

// boardPanelInputs 把提交的面板转成 biz 对象；布局与 URL 校验都在 biz。
func boardPanelInputs(panels []*v1.BoardPanelInput) []bo.BoardPanel {
	out := make([]bo.BoardPanel, 0, len(panels))
	for _, panel := range panels {
		out = append(out, bo.BoardPanel{
			UID: panel.GetUid(), Title: panel.GetTitle(), Description: panel.GetDescription(),
			GrafanaURL: panel.GetGrafanaUrl(), Width: int(panel.GetWidth()), Height: int(panel.GetHeight()),
		})
	}
	return out
}

func boardItem(item bo.Board) *v1.BoardItem {
	panels := make([]*v1.BoardPanelItem, 0, len(item.Panels))
	for _, panel := range item.Panels {
		panels = append(panels, &v1.BoardPanelItem{
			Uid: panel.UID, Title: panel.Title, Description: panel.Description,
			GrafanaUrl: panel.GrafanaURL, Width: int32(panel.Width), Height: int32(panel.Height),
			Sort: int32(panel.Sort),
		})
	}
	return &v1.BoardItem{
		Uid: item.UID, Name: item.Name, Description: item.Description, Status: item.Status,
		Panels: panels, CreatedAt: formatTime(item.CreatedAt), UpdatedAt: formatTime(item.UpdatedAt),
	}
}

func boardListItem(item bo.Board) *v1.BoardListItem {
	return &v1.BoardListItem{
		Uid: item.UID, Name: item.Name, Description: item.Description, Status: item.Status,
		PanelCount: int32(item.PanelCount), CreatedAt: formatTime(item.CreatedAt), UpdatedAt: formatTime(item.UpdatedAt),
	}
}

// formatTime 按项目约定输出 RFC3339；零值返回空字符串。
func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
