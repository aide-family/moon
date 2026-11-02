/** 大盘状态，与后端约定为大写英文。 */
export type BoardStatus = "ACTIVE" | "DISABLED"

/** 详情里的单个面板，sort 由服务端按提交顺序从 0 起连续重写。 */
export interface BoardPanelItem {
  uid: string
  title: string
  description: string
  /** Grafana 分享 / 嵌入地址，浏览器直接加载，不经过控制面。 */
  grafanaUrl: string
  /** 占 12 栏栅格中的几列。 */
  width: number
  /** iframe 内容区像素高，不含标题与描述。 */
  height: number
  sort: number
}

/** 创建 / 更新时提交的面板参数：uid 为空表示新增。 */
export interface BoardPanelInput {
  uid?: string
  title: string
  description: string
  grafanaUrl: string
  width: number
  height: number
}

/** 大盘详情，面板按展示顺序返回。 */
export interface BoardItem {
  uid: string
  name: string
  description: string
  status: BoardStatus
  panels: BoardPanelItem[]
  createdAt: string
  updatedAt: string
}

/** 列表项不返回面板，只带面板数量。 */
export interface BoardListItem extends Omit<BoardItem, "panels"> {
  panelCount: number
}

export interface ListBoardsParams {
  page?: number
  pageSize?: number
  name?: string
  /** 空串表示不限制状态。 */
  status?: BoardStatus | ""
}

export interface ListBoardsResult {
  items: BoardListItem[]
  total: number
  page: number
  pageSize: number
}

/** 新建与整体更新共用；更新时 uid 必填，面板列表整体替换。 */
export interface SaveBoardParams {
  uid?: string
  name: string
  description: string
  panels: BoardPanelInput[]
}
