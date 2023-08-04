import {Post} from "@/apis/requests";
import type {GroupListReply} from "@/apis/prom/group/group";

const API = "/prom/v1/group/list"

async function GroupList() {
    return Post<GroupListReply>(API, {
        "page": {
            "size": 10,
            "current": 1
        }
    })
}

export default GroupList