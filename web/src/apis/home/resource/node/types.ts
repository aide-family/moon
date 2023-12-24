import { Map, PageReq, PageRes, Response } from '@/apis/types'
export type NodeEditReq = {
    name: string
    cname: string
    band_width: number
    status: string
    space_instance: string
    type: string
    purpose: string
    customer: string
    remark: string
    agent_group?: string
}

export type NodeItem = {
    created_by: string
    created_at: number
    updated_by: string
    updated_at: number
    id: string
    space_instance: string
    name: string
    cname: string
    band_width: number
    status?: Map
    supplier_chat: string
    customer_chat: string
    type?: Map
    purpose?: Map
    is_monitor: number
    is_jump: number
    customer?: Map
    ikuai?: string
    remark: string
    resource_supplier?: Map
}

export type NodeListReq = PageReq & {
    name?: string
}

export type NodeListRes = Response<PageRes<NodeItem>>
