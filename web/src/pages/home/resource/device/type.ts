import { Map } from '@/apis/types'

export type DeviceItemType = {
    deleted?: number
    revision?: null
    created_by?: string
    created_at?: number
    updated_by?: string
    updated_at?: number
    id?: string
    space_instance?: string
    node?: Map
    type?: Map
    supplier?: Map
    host_name?: string
    sn?: string
    source?: Map
    status?: Map
    remark?: string
    ipmi?: string | null
    manage_ip?: string | null
    manage_port?: string | null
}
