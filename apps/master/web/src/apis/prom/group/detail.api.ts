import {Get} from "@/apis/requests";
import {GroupDetailReply} from "@/apis/prom/group/group";

const API = "/prom/v1/group/detail"

async function GroupDetail(id: number) {
    return  Get<GroupDetailReply>(API+"/"+id)
}

export default GroupDetail