import { Map } from '@/apis/types'

export type AccountItemType = {
    created_by: string
    created_at: number
    updated_by: string
    updated_at: number
    id: string
    space_instance: string
    type: Map
    name: string
    password: string
    ipv4: string
    ipv6: string
    vlan_id: number
    vlan_type: Map
    bandwidth: number
    status: Map
    node: Map
    host: Map
    switch_device: Map
    switch_port: string
}
