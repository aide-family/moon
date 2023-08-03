import {Get} from "@/apis/requests";
import {NodeDetailReply} from "@/apis/prom/node/node";

const API = "/prom/v1/node/detail"

async function NodeDetail(id: number) {
    return  Get<NodeDetailReply>(API+"/"+id)
}

export default NodeDetail