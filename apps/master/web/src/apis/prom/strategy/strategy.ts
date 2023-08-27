import type { ListReponse, M, Query, Response } from "@/apis/type";
import type {
  PromStrategyItem,
  PromStrategyItemRequest,
} from "@/apis/prom/prom";

// 规则列表请求参数
export type ListStrategyRequest = {
  query: Query;
  strategy?: PromStrategyItemRequest;
};

export const defaultListStrategyRequest: ListStrategyRequest = {
  query: {
    page: {
      current: 1,
      size: 10,
    },
    sort: [
      {
        field: "updated_at",
        asc: false,
      },
    ],
  },
};

// 规则列表响应参数
export type ListStrategyReply = ListReponse & {
  strategies: PromStrategyItem[];
};

// 规则详情响应参数
export type StrategyDetailReply = Response & {
  strategy: PromStrategyItem;
};

// 规则创建参数
export type StrategyCreateItem = {
  // 规则所属规则组ID
  groupId: number;
  // 规则名称
  alert: string;
  // 规则PromQL表达式
  expr: string;
  // 持续时间, 单位(s|m|h|d), 表示prom ql表达式持续多久时间满足条件
  for: string;
  // 规则标签, 用于自定义告警标签信息
  labels: M;
  // 规则注释, 用于自定义告警注释信息, 例如:告警标题、告警内容等
  annotations: M;
  // 规则tag id, 用于标记规则属性, 例如告警对象领域, 业务领域等, 与categories一一对应
  categorieIds: number[];
  // 告警等级ID, 用于标记告警等级, 每一个告警规则必有一个告警等级
  alertLevelId: number;
  // 告警页面, 用于标记告警页面, 告警发生时, 告警信息会发送到对应的告警页面
  alarmPageIds: number[];
};
