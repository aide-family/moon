import type {PageReply, Resp} from "@/apis/requests";
import type {NodeItem} from "@/apis/prom/prom";

export type NodeListReply = {
    page: PageReply
    response: Resp
    list: NodeItem[]
}

export type NodeDetailReply = {
    response: Resp
    node: NodeItem
}