package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// BoardRepo 是大盘与面板的仓储。面板随大盘整体替换，不单独提供 CRUD。
type BoardRepo interface {
	CreateBoard(ctx context.Context, in bo.Board) (*bo.Board, error)
	UpdateBoard(ctx context.Context, in bo.Board) (*bo.Board, error)
	UpdateBoardStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Board, error)
	DeleteBoard(ctx context.Context, namespaceUID, uid string) error
	GetBoard(ctx context.Context, namespaceUID, uid string) (*bo.Board, error)
	ListBoards(ctx context.Context, namespaceUID string, filter bo.BoardListFilter, page, pageSize int) ([]bo.Board, int, error)
}
