import { Map, PageReq, PageRes, Response } from '@/apis/types'
export type EditAccountReq = {
    space_instance: string
    type: string
    name: string
    password: string
    ipv4: string
    ipv6: string
    vlan_id: string
    vlan_type: string
    bandwidth: number
    status: string
    node: string
    host: string
    switch_device: string
    switch_port: string
}

export type AccountItem = {
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

export type AccountListReq = PageReq & {
    name?: string
    type?: string
}

export type AccountListRes = Response<PageRes<AccountItem>>
