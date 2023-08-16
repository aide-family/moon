import {Delete, Get, Post, Put} from "@/apis/requests";
import type {Response} from "@/apis/type";
import type {
    ListStrategyReply,
    ListStrategyRequest,
    StrategyCreateItem,
    StrategyDetailReply
} from "@/apis/prom/strategy/strategy";
import {Status, StatusMap} from "@/apis/prom/prom";

/**
 * 规则列表请求参数, 复数形式, 模块+版本+模型复数
 * @list: POST /prom/v1/strategies
 * @status: POST /prom/v1/strategies/status
 */
const StrategyListAPI = "/prom/v1/strategies"

/**
 * 规则单一操作, 单数形式
 * @create: POST /prom/v1/strategy
 * @update: POST /prom/v1/strategy/{id}
 * @delete: POST /prom/v1/strategy/{id}
 * @detail: POST /prom/v1/strategy/{id}
 */
const StrategyAPI = "/prom/v1/strategy"

/**
 * 规则列表
 * @param params 查询参数
 * @constructor
 */
function StrategyList(params: ListStrategyRequest) {
    return Post<ListStrategyReply>(StrategyListAPI, params || {})
}

/**
 * 规则详情
 * @param id 规则ID
 * @constructor
 */
function StrategyDetail(id: number) {
    return Get<StrategyDetailReply>([StrategyAPI, id].join("/"))
}

/**
 * 创建规则
 * @param params 创建参数
 * @constructor
 */
function StrategyCreate(params: StrategyCreateItem) {
    return Post<Response>(StrategyAPI, {strategy: params})
}

/**
 * 规则修改
 * @param id 规则ID
 * @param params 规则修改参数
 * @constructor
 */
function StrategyUpdate(id: number, params: StrategyCreateItem) {
    return Put<Response>([StrategyAPI, id].join("/"), {strategy: params})
}

/**
 * 规则删除
 * @param id 规则ID
 * @constructor
 */
function StrategyDelete(id: number) {
    return Delete<Response>([StrategyAPI, id].join("/"))
}

/**
 * 规则状态修改
 * @param ids 规则ID列表
 * @param status 修改的状态
 * @constructor
 */
function StrategyUpdateStatus(ids: number[], status: Status) {
    return Put<Response>([StrategyListAPI, "status"].join("/"), {
        status: StatusMap[status].opposite?.number,
        ids: ids
    })
}

export {
    StrategyList,
    StrategyDetail,
    StrategyCreate,
    StrategyUpdate,
    StrategyDelete,
    StrategyUpdateStatus
}