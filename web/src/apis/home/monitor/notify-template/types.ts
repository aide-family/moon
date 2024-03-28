import { NotifyTemplateType, PageReqType, PageResType } from '@/apis/types'

/**
 * // 创建模板请求参数
message CreateTemplateRequest {
  // 模板内容
  string content = 2 [(validate.rules).string.min_len = 1];
  // 所属策略
  uint32 strategyId = 3 [(validate.rules).uint32.gt = 0];
  // 模板类型
  NotifyTemplateType notifyType = 4 [(validate.rules).enum.defined_only = true];
}
 */
export interface CreateTemplateRequest {
    content: string
    strategyId: number
    notifyType: NotifyTemplateType
}

/**
 * // 更新模板请求参数
message UpdateTemplateRequest {
  uint32 id = 1 [(validate.rules).uint32.gt = 0];
  // 模板内容
  string content = 2 [(validate.rules).string.min_len = 1];
  // 所属策略
  uint32 strategyId = 3 [(validate.rules).uint32.gt = 0];
  // 模板类型
  NotifyTemplateType notifyType = 4 [(validate.rules).enum.defined_only = true];
}
 */
export interface UpdateTemplateRequest extends CreateTemplateRequest {
    id: number
}

/**
 * // 通知模板
message NotifyTemplateItem {
  // 模板ID
  uint32 id = 1;
  // 模板内容
  string content = 2;
  // 所属策略
  uint32 strategyId = 3;
  // 模板类型
  int32 notifyType = 4;
}
 */
export interface NotifyTemplateItem {
    id: number
    content: string
    strategyId: number
    notifyType: NotifyTemplateType
}

/**
 * message GetTemplateReply {
  NotifyTemplateItem detail = 1;
}
 */
export interface GetTemplateReply {
    detail: NotifyTemplateItem
}

/**
 * // 获取模板列表请求参数
message ListTemplateRequest {
  PageRequest page = 1 [(validate.rules).message.required = true];
  uint32 strategyId = 2 [(validate.rules).uint32.gt = 0];
}
 */
export interface ListTemplateRequest {
    page: PageReqType
    strategyId: number
}

/**
 * message ListTemplateReply {
  PageReply page = 1;
  repeated NotifyTemplateItem list = 2;
}
 */
export interface ListTemplateReply {
    page: PageResType
    list: NotifyTemplateItem[]
}
