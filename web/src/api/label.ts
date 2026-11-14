import { get } from "@/api/http"

export interface LabelOption {
  key: string
  values: string[]
}

export interface ListLabelOptionsResult {
  items: LabelOption[]
}

/** GET /v1/label-options 返回当前租户各配置对象出现过的标签键值集合。 */
export function listLabelOptions(): Promise<ListLabelOptionsResult> {
  return get<ListLabelOptionsResult>("/label-options")
}
