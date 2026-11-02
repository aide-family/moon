package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/board"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/boardpanel"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

// CreateBoard 在同一个事务里写大盘与面板。面板 UID 由服务端生成，提交顺序即展示顺序。
func (r *boardRepo) CreateBoard(ctx context.Context, in bo.Board) (*bo.Board, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.Board.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDescription(in.Description).
		SetStatus(board.StatusACTIVE).
		SetCreatedBy(in.CreatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := createBoardPanels(ctx, tx, in.NamespaceUID, row.UID, in.Panels); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetBoard(ctx, in.NamespaceUID, row.UID)
}

// UpdateBoard 整体替换面板：保留提交里带 UID 的面板，删除不再出现的面板。
func (r *boardRepo) UpdateBoard(ctx context.Context, in bo.Board) (*bo.Board, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.Board.Update().
		Where(board.UIDEQ(in.UID), board.NamespaceUIDEQ(in.NamespaceUID), board.DeletedAtIsNil()).
		SetName(in.Name).
		SetDescription(in.Description).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("board not found")
	}
	if err := replaceBoardPanels(ctx, tx, in.NamespaceUID, in.UID, in.Panels); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetBoard(ctx, in.NamespaceUID, in.UID)
}

func (r *boardRepo) UpdateBoardStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Board, error) {
	n, err := r.client.Board.Update().
		Where(board.UIDEQ(uid), board.NamespaceUIDEQ(namespaceUID), board.DeletedAtIsNil()).
		SetStatus(board.Status(status)).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("board not found")
	}
	return r.GetBoard(ctx, namespaceUID, uid)
}

// DeleteBoard 软删除大盘及其面板。面板不再单独恢复。
func (r *boardRepo) DeleteBoard(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC()
	n, err := tx.Board.Update().
		Where(board.UIDEQ(uid), board.NamespaceUIDEQ(namespaceUID), board.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("board not found")
	}
	if _, err := tx.BoardPanel.Update().
		Where(boardpanel.NamespaceUIDEQ(namespaceUID), boardpanel.BoardUIDEQ(uid), boardpanel.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *boardRepo) GetBoard(ctx context.Context, namespaceUID, uid string) (*bo.Board, error) {
	row, err := r.client.Board.Query().Where(
		board.UIDEQ(uid), board.NamespaceUIDEQ(namespaceUID), board.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	panels, err := r.client.BoardPanel.Query().
		Where(boardpanel.NamespaceUIDEQ(namespaceUID), boardpanel.BoardUIDEQ(uid), boardpanel.DeletedAtIsNil()).
		Order(ent.Asc(boardpanel.FieldSort), ent.Asc(boardpanel.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toBoard(row, panels)
	return &item, nil
}

func (r *boardRepo) ListBoards(ctx context.Context, namespaceUID string, filter bo.BoardListFilter, page, pageSize int) ([]bo.Board, int, error) {
	query := r.client.Board.Query().Where(board.NamespaceUIDEQ(namespaceUID), board.DeletedAtIsNil())
	if filter.Name != "" {
		query = query.Where(board.NameContainsFold(filter.Name))
	}
	if filter.Status != "" {
		query = query.Where(board.StatusEQ(board.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(board.FieldUpdatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Board, 0, len(rows))
	if len(rows) == 0 {
		return items, total, nil
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.UID)
	}
	counts, err := r.boardPanelCounts(ctx, namespaceUID, uids)
	if err != nil {
		return nil, 0, err
	}
	for _, row := range rows {
		item := toBoard(row, nil)
		item.PanelCount = counts[row.UID]
		items = append(items, item)
	}
	return items, total, nil
}

// boardPanelCounts 一次取出当前页大盘的面板数量，避免逐个大盘查询。
func (r *boardRepo) boardPanelCounts(ctx context.Context, namespaceUID string, boardUIDs []string) (map[string]int, error) {
	rows, err := r.client.BoardPanel.Query().
		Where(
			boardpanel.NamespaceUIDEQ(namespaceUID),
			boardpanel.BoardUIDIn(boardUIDs...),
			boardpanel.DeletedAtIsNil(),
		).
		Select(boardpanel.FieldBoardUID).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	counts := make(map[string]int, len(boardUIDs))
	for _, row := range rows {
		counts[row.BoardUID]++
	}
	return counts, nil
}

func createBoardPanels(ctx context.Context, tx *ent.Tx, namespaceUID, boardUID string, panels []bo.BoardPanel) error {
	for index, panel := range panels {
		if _, err := tx.BoardPanel.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetBoardUID(boardUID).
			SetTitle(panel.Title).
			SetDescription(panel.Description).
			SetGrafanaURL(panel.GrafanaURL).
			SetWidth(panel.Width).
			SetHeight(panel.Height).
			SetSort(index).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// replaceBoardPanels 更新已有面板、创建新面板、软删除被移除的面板，提交顺序写入 sort。
func replaceBoardPanels(ctx context.Context, tx *ent.Tx, namespaceUID, boardUID string, panels []bo.BoardPanel) error {
	existing, err := tx.BoardPanel.Query().
		Where(boardpanel.NamespaceUIDEQ(namespaceUID), boardpanel.BoardUIDEQ(boardUID), boardpanel.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return err
	}
	kept := make(map[string]struct{}, len(panels))
	for index, panel := range panels {
		if panel.UID == "" {
			if _, err := tx.BoardPanel.Create().
				SetUID(id.New()).
				SetNamespaceUID(namespaceUID).
				SetBoardUID(boardUID).
				SetTitle(panel.Title).
				SetDescription(panel.Description).
				SetGrafanaURL(panel.GrafanaURL).
				SetWidth(panel.Width).
				SetHeight(panel.Height).
				SetSort(index).
				Save(ctx); err != nil {
				return err
			}
			continue
		}
		kept[panel.UID] = struct{}{}
		n, err := tx.BoardPanel.Update().
			Where(
				boardpanel.UIDEQ(panel.UID),
				boardpanel.NamespaceUIDEQ(namespaceUID),
				boardpanel.BoardUIDEQ(boardUID),
				boardpanel.DeletedAtIsNil(),
			).
			SetTitle(panel.Title).
			SetDescription(panel.Description).
			SetGrafanaURL(panel.GrafanaURL).
			SetWidth(panel.Width).
			SetHeight(panel.Height).
			SetSort(index).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			// 面板 UID 必须属于当前大盘；不能借 UID 改别的面板。
			return merr.Invalid("panel not found")
		}
	}
	now := time.Now().UTC()
	for _, row := range existing {
		if _, ok := kept[row.UID]; ok {
			continue
		}
		if _, err := tx.BoardPanel.Update().
			Where(boardpanel.IDEQ(row.ID), boardpanel.DeletedAtIsNil()).
			SetDeletedAt(now).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func toBoard(row *ent.Board, panels []*ent.BoardPanel) bo.Board {
	item := bo.Board{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Description: row.Description, Status: string(row.Status), CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	item.Panels = make([]bo.BoardPanel, 0, len(panels))
	for _, panel := range panels {
		item.Panels = append(item.Panels, bo.BoardPanel{
			UID: panel.UID, NamespaceUID: panel.NamespaceUID, BoardUID: panel.BoardUID,
			Title: panel.Title, Description: panel.Description, GrafanaURL: panel.GrafanaURL,
			Width: panel.Width, Height: panel.Height, Sort: panel.Sort,
		})
	}
	return item
}
