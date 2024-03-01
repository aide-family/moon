import { POST } from '@/apis/request'
import { LogListReq, LogListResp } from './types'

enum URL {
    /** 获取日志列表 */
    LIST = '/api/v1/syslog/list'
}

/** 获取日志列表 */
const list = (data: LogListReq) => {
    return POST<LogListResp>(URL.LIST, data)
}

export const logApi = {
    list
}
