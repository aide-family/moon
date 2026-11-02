import { del, get, post, put } from "@/api/http"

import type {
  BoardItem,
  BoardStatus,
  ListBoardsParams,
  ListBoardsResult,
  SaveBoardParams,
} from "./types"

/** GET /v1/boards */
export function listBoards(
  params: ListBoardsParams = {}
): Promise<ListBoardsResult> {
  return get<ListBoardsResult>("/boards", params)
}

/** GET /v1/board/{uid} */
export function getBoard(uid: string): Promise<{ board: BoardItem }> {
  return get<{ board: BoardItem }>(`/board/${uid}`)
}

/** POST /v1/board */
export function createBoard(
  body: SaveBoardParams
): Promise<{ board: BoardItem }> {
  return post<{ board: BoardItem }>("/board", body)
}

/** PUT /v1/board/{uid}：面板整体替换，uid 为空的面板视为新增。 */
export function updateBoard(
  body: SaveBoardParams & { uid: string }
): Promise<{ board: BoardItem }> {
  return put<{ board: BoardItem }>(`/board/${body.uid}`, body)
}

/** PUT /v1/board/{uid}/status：只切换启停状态，不重写面板。 */
export function updateBoardStatus(
  uid: string,
  status: BoardStatus
): Promise<{ board: BoardItem }> {
  return put<{ board: BoardItem }>(`/board/${uid}/status`, { uid, status })
}

/** DELETE /v1/board/{uid} */
export function deleteBoard(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/board/${uid}`)
}

export type {
  BoardItem,
  BoardListItem,
  BoardPanelInput,
  BoardPanelItem,
  BoardStatus,
  ListBoardsParams,
  ListBoardsResult,
  SaveBoardParams,
} from "./types"
