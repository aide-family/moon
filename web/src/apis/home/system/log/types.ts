import {
    ModuleType,
    PageReqType,
    PageResType,
    SysLogActionType
} from '@/apis/types'
import { UserSelectItem } from '../user/types'

interface LogItem {
    title: string
    content: string
    moduleId: number
    moduleName: string
    createdAt: number
    user?: UserSelectItem
    action: SysLogActionType
}

interface LogListReq {
    page: PageReqType
    moduleId: number
    moduleName: ModuleType
}

interface LogListResp {
    list: LogItem[]
    page: PageResType
}

export const defauleLogListReq: LogListReq = {
    page: {
        curr: 1,
        size: 10
    },
    moduleId: 0,
    moduleName: ModuleType.ModelTypeOther
}

export type { LogItem, LogListReq, LogListResp }
