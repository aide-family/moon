import type { ListReponse, Query, Response } from "@/apis/type";
import type {
  DatasourceItem,
  PromDict,
  PromDictRequest,
} from "@/apis/prom/prom";
import { Category } from "@/apis/prom/prom";

export type ListDictRequest = {
  query: Query;
  dict?: PromDictRequest;
};

export const defaultListDictRequest: ListDictRequest = {
  query: {
    page: {
      current: 1,
      size: 10,
    },
  },
};

export type ListDictReply = ListReponse & {
  dicts: PromDict[];
};

export type ListDatasourceReply = Response & {
  datasources: DatasourceItem[];
};

export type DictDetailReply = Response & {
  dict: PromDict;
};

export type DictCreateItem = {
  // 字典名称
  name: string;
  // 字典备注
  remark?: string;
  // 字典类别
  category: Category;
  // 字典颜色
  color?: string;
};
