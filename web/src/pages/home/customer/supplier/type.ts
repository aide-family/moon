import { Map } from '@/apis/types'

export type SupplierItemType = {
    created_by?: string
    created_at?: number
    updated_by?: string
    updated_at?: number
    id?: string
    space_instance?: string
    type?: Map
    company?: string
    contacts?: string
    phone?: string
    wechat?: string
    address?: string
}
