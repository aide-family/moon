import type { PageReply, PageReq, Response } from "@/apis/type";
import type { AlarmPage } from "@/apis/prom/prom";

export type ListSimpleAlarmPageRequest = {
  page: PageReq;
  keyword?: string;
};

export type ListSimpleAlarmPageReply = Response & {
  page: PageReply;
  alarmPages: AlarmPage[];
};
