import type { Query } from "@/apis/type";

export type AlarmItem = {
  id: number;
};

export type AlarmQueryParams = {
  query?: Query;
  strategyId?: number;
  pageId?: number;
  levelId?: number;
  duration?: number;
  node?: string;
  startAt?: number;
  endAt?: number;
};
