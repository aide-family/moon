import { Post } from "@/apis/requests";
import {
  ListSimpleAlarmPageReply,
  ListSimpleAlarmPageRequest,
} from "@/apis/prom/alarm/alarm-page";

const AlarmPageListAPI = "/prom/v1/alarm-pages";

function AlarmPageSimpleList(params: ListSimpleAlarmPageRequest) {
  return Post<ListSimpleAlarmPageReply>(
    [AlarmPageListAPI, "simple"].join("/"),
    params || {}
  );
}

export { AlarmPageSimpleList };
