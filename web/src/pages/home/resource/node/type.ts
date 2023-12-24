import { Map } from '@/apis/types'

export type NodeItemType = {
    created_by?: string
    created_at?: number
    updated_by?: string
    updated_at?: number
    id?: string
    space_instance?: string
    name?: string
    cname?: string
    band_width?: number
    status?: Map
    supplier_chat?: string
    customer_chat?: string
    type?: Map
    purpose?: Map
    is_monitor?: number
    is_jump?: number
    customer?: Map
    ikuai?: string
    remark?: string
    resource_supplier?: Map
    agent_group?: string
}
