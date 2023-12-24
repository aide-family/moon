import { GET, POST, PostForm } from '@/apis/request'
import {
    AccountItem,
    AccountListReq,
    AccountListRes,
    EditAccountReq
} from './types'
import { Callback, Response } from '@/apis/types'
import { getSpaceID } from '@/context'

enum URL {
    accountPage = '/assets/pcdn/account/v1/select/page',
    accountDetail = '/assets/pcdn/account/v1/select/like',
    accountInsert = '/assets/pcdn/account/v1/insert',
    accountUpdate = '/assets/pcdn/account/v1/update',
    accountDelete = '/assets/pcdn/account/v1/delete'
}

const GetAccountList = async (params: AccountListReq, call?: Callback) => {
    call?.setLoading?.(true)
    const res = await GET<AccountListRes>(URL.accountPage, {
        ...params,
        space_instance: getSpaceID()
    })
    call?.setLoading?.(false)
    if (res?.code !== 0) {
        call?.ERROR?.(res?.message)
        return Promise.reject(res?.message)
    }
    call?.OK?.(res.data)
    return res.data
}

const GetAccountDetail = async (id: string, call?: Callback) => {
    call?.setLoading?.(true)
    const res = await GET<Response<AccountItem[]>>(URL.accountDetail, {
        id,
        space_instance: getSpaceID()
    })
    call?.setLoading?.(false)
    if (res.code !== 0) {
        call?.ERROR?.(res.message)
        return Promise.reject(res.message)
    }

    if (res.data.length === 0) {
        call?.ERROR?.('账户不存在')
        return Promise.reject('账户不存在')
    }

    if (res.data.length > 1) {
        call?.ERROR?.('账户数据异常')
        return Promise.reject('账户数据异常')
    }

    call?.OK?.(res.data[0])
    return res.data[0]
}

const AddAccount = async (params: EditAccountReq, call?: Callback) => {
    call?.setLoading?.(true)
    const res = await POST<Response>(URL.accountInsert, {
        ...params,
        space_instance: getSpaceID()
    })
    call?.setLoading?.(false)
    if (res.code !== 0) {
        call?.ERROR?.(res.message)
        return Promise.reject(res?.message)
    }
    call?.OK?.(res)
    return res
}

const UpdateAccount = async (
    id: string,
    params: EditAccountReq,
    call?: Callback
) => {
    call?.setLoading?.(true)
    const res = await POST<Response>(URL.accountUpdate, {
        ...params,
        id,
        space_instance: getSpaceID()
    })
    call?.setLoading?.(false)
    if (res.code !== 0) {
        call?.ERROR?.(res.message)
        return Promise.reject(res?.message)
    }
    call?.OK?.(res)
    return res
}

export const DeleteAccount = async (id: string) => {
    const res = await PostForm<Response>(URL.accountDelete, {
        id,
        space_instance: getSpaceID()
    })
    if (res?.code !== 0) {
        return Promise.reject(res?.message)
    }
    return res
}

export { GetAccountList, GetAccountDetail, AddAccount, UpdateAccount }
