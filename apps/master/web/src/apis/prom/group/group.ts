import type {PageReply, Resp} from "@/apis/requests";
import type {GroupItem} from "@/apis/prom/prom";

export type GroupListReply = {
    page: PageReply
    response: Resp
    list: GroupItem[]
}

export type GroupDetailReply = {
    response: Resp
    group: GroupItem
}