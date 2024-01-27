import { GET, POST, PostForm } from '@/apis/request'
import { NodeEditReq, NodeItem, NodeListReq, NodeListRes } from './types'
import { Callback, PageRes, Response } from '@/apis/types'

enum URL {
    nodePage = '/assets/node/v1/select/page',
    nodeAdd = '/assets/node/v1/insert',
    nodeUpdate = '/assets/node/v1/update',
    nodeDetail = '/assets/node/v1/select/like',
    nodeDelete = '/assets/node/v1/delete'
}

const GetNodeList = async (
    params: NodeListReq,
    call?: Callback<PageRes<NodeItem>, string>
) => {
    call?.setLoading?.(true)
    const res = await GET<NodeListRes>(URL.nodePage, params)
    call?.setLoading?.(false)
    if (res?.code !== 0) {
        call?.ERROR?.(res?.message)
        return Promise.reject(res?.message)
    }
    call?.OK?.(res?.data)
    return res.data
}

const AddNode = async (params: NodeEditReq) => {
    const res = await POST<Response>(URL.nodeAdd, {
        ...params
    })
    if (res?.code !== 0) {
        return Promise.reject(res?.message)
    }
    return res.data
}

const UpdateNode = async (id: string, params: NodeEditReq) => {
    const res = await POST<Response>(URL.nodeUpdate, {
        id,
        ...params
    })
    if (res?.code !== 0) {
        return Promise.reject(res?.message)
    }
    return res.data
}

const GetNodeDetail = async (id: string, call?: Callback) => {
    call?.setLoading?.(true)
    const res = await GET<Response<NodeItem[]>>(URL.nodeDetail, { id })
    call?.setLoading?.(false)
    if (res.code !== 0) {
        call?.ERROR?.(res.message)
        return Promise.reject(res.message)
    }

    if (res.data.length === 0) {
        call?.ERROR?.('节点不存在')
        return Promise.reject('节点不存在')
    }

    if (res.data.length > 1) {
        call?.ERROR?.('节点数据异常')
        return Promise.reject('节点数据异常')
    }
    call?.OK?.(res.data[0])

    return res.data[0]
}

const DeleteNode = async (id: string) => {
    const res = await PostForm<Response>(URL.nodeDelete, {
        id
    })

    if (res?.code !== 0) {
        return Promise.reject(res?.message)
    }
    return res
}

export { GetNodeList, AddNode, UpdateNode, GetNodeDetail, DeleteNode }
