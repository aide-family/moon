import {Post} from "@/apis/requests";
import type {NodeListReply} from "@/apis/prom/node/node";

const API = "/prom/v1/node/list"

async function NodeList() {
    return Post<NodeListReply>(API, {
        "page": {
            "size": 10,
            "current": 1
        }
    })
}

export default NodeList

