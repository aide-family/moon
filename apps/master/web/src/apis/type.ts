// 自定义map类型
export type M<T = string> = { [key: string]: T };

// 分页响应参数
export type PageReply = {
  // 查询的页码
  current: number;
  // 每页条数
  size: number;
  // 总条数
  total: string;
};

// 列表返回字段响应参数
export type Field = {
  // 字段key, 一般为字段名称
  field: string;
  // 字段名称, 一般为中文名称
  label: string;
};

// 基础响应参数
export type BaseResp = {
  // 响应码, 0为成功
  code: string;
  // 响应消息
  message: string;
  // 响应元数据
  metadata?: M;
};

// 基础响应格式
export type Response = {
  // 基础响应参数
  response: BaseResp;
};

// 列表响应参数
export type ListReponse = Response & {
  result: Result;
};

// 列表响应额外补充信息
export type Result = {
  // 分页数据
  page: PageReply;
  // 列表字段信息列表
  fields: Field[];
};

// 分页请求参数
export type PageReq = {
  // 当前页码
  current: number;
  // 每页条数
  size: number;
};

// 列表排序请求参数
export type Sort = {
  // 排序字段
  field: string;
  // 是否升序
  asc: boolean;
};

// 列表查询请求参数
export type Query = {
  // 分页参数
  page: PageReq;
  // 排序字段列表
  sort?: Sort[];
  // 查询字段列表
  fields?: string[];
  // 查询关键字
  keyword?: string;
  // 查询起始时间
  startAt?: string;
  // 查询结束时间
  endAt?: string;
  // 查询的时间字段
  timeField?: string;
};

export const defaultPage: PageReq = {
  current: 1,
  size: 10,
};
