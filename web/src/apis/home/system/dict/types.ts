import { Category, PageReqType, PageResType, Status } from '@/apis/types'

/**创建字典 */
interface CreateDict {
    name: string
    category: number
    color: string
    remark: string
}
/**修改字典 */
interface UpdateDict {
    id: number
    name: string
    category: number
    color: string
    remark: string
    status: number
}
/**根据Id获取数据 */
interface DictById {
    id: number
}
/**获取字典列表参数 */
interface DictListReq {
    page: PageReqType
    keyword?: string
    category?: number
}

/**字典列表返回数据 */
interface DictListRes {
    list: DictListItem[]
    page: PageResType
}

/**字典列表项 */
interface DictListItem {
    id: number
    name: string
    category: Category
    color: string
    status: Status
    remark: string
    createdAt: number
    updatedAt: number
    deletedAt: number
}

/**获取字典详情参数 */
interface DictDetailReq {
    id: number
    isDeleted: boolean
}
/**字典详情返回数据 */
interface DictDetailRes {
    promDict: DictListItem
}
/** 获取字典列表, 用于下拉选择*/
interface DictSelectReq {
    isDeleted?: boolean
    page: PageReqType
    keyword?: string
    category?: number
}
/**获取字典列表, 用于下拉选择返回数据 */
interface DictSelectRes {
    list: DictSelectItem[]
    page: PageResType
}

/**字典下拉选择项 */
interface DictSelectItem {
    value: number
    label: string
    category: number
    color: string
    status: number
    remark: string
    isDeleted: false
}
interface dictBatchDeleteType {
    ids: number[]
}
interface DictBatchUpdateStatusType {
    ids: number[]
    status: Status
}

export type {
    CreateDict,
    UpdateDict,
    DictById,
    DictListReq,
    DictListRes,
    DictListItem,
    DictDetailReq,
    DictDetailRes,
    DictSelectReq,
    DictSelectRes,
    DictSelectItem,
    dictBatchDeleteType,
    DictBatchUpdateStatusType
}
