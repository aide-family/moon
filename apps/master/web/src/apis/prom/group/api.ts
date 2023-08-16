import {Delete, Get, Post, Put} from "@/apis/requests";
import type {GroupCreateItem, GroupDetailReply, ListGroupReply, ListGroupRequest} from "@/apis/prom/group/group";
import {Status, StatusMap} from "@/apis/prom/prom";
import type {Response} from "@/apis/type";

/**
 * 规则组列表请求参数, 复数形式, 模块+版本+模型复数, 为了支持复杂的查询条件, 使用POST请求
 * @list: POST /prom/v1/groups
 */
const GroupListAPI = "/prom/v1/groups"

/**
 * 规则单一操作, 单数形式, 模块+版本+模型单数, 对应操作使用不同的HTTP请求方法, 查询GET, 创建POST, 更新PUT, 删除DELETE
 * @create: POST /prom/v1/group
 * @update: PUT /prom/v1/group/{id}
 * @delete: DELETE /prom/v1/group/{id}
 * @detail: GET /prom/v1/group/{id}
 */
const GroupAPI = "/prom/v1/group"

/**
 * 获取规则组列表
 * @param params 查询参数
 */
function GroupList(params: ListGroupRequest) {
    return Post<ListGroupReply>(GroupListAPI, params || {})
}

/**
 * 获取规则组详情
 * @param id 规则组ID
 */
function GroupDetail(id: number) {
    return Get<GroupDetailReply>([GroupAPI, id].join("/"))
}

/**
 * 创建规则组
 * @param params 规则组参数
 */
function GroupCreate(params: GroupCreateItem) {
    return Post<Response>(GroupAPI, {group: params})
}

/**
 * 更新规则组
 * @param id 规则组ID
 * @param params 规则组参数
 */
function GroupUpdate(id: number, params: GroupCreateItem) {
    return Put<Response>([GroupAPI, id].join("/"), {group: params})
}

/**
 * 删除规则组
 * @param id 规则组ID
 */
function GroupDelete(id: number) {
    return Delete<Response>([GroupAPI, id].join("/"))
}

/**
 * 更新规则组状态
 * @param ids 规则组ID
 * @param status 规则组状态
 */
function GroupUpdatesStatus(ids: number[], status: Status) {
    return Put<Response>([GroupListAPI, "status"].join("/"), {status: StatusMap[status].opposite?.number, ids: ids})
}

export {
    GroupList,
    GroupDetail,
    GroupCreate,
    GroupUpdate,
    GroupDelete,
    GroupUpdatesStatus,
}