import { GET, POST, PostForm } from '@/apis/request'
import { EquipmentAddReq, EquipmentListItem, EquipmentListReq } from './types'
import { Callback, PageRes, Response } from '@/apis/types'

enum URL {
    equipmentPage = '/assets/equipment/v1/select/page',
    equipmentAdd = '/assets/equipment/v1/insert',
    equipmentUpdate = '/assets/equipment/v1/update',
    equipmentDetail = '/assets/equipment/v1/select/like',
    equipmentDelete = '/assets/equipment/v1/delete'
}

const GetEquipmentList = async (
    params: EquipmentListReq,
    call?: Callback<PageRes<EquipmentListItem>, string>
) => {
    call?.setLoading?.(true)
    const response = await GET<Response<PageRes<EquipmentListItem>>>(
        URL.equipmentPage,
        params
    )
    call?.setLoading?.(false)
    if (response?.code !== 0) {
        call?.ERROR?.(response?.message)
        return Promise.reject(response?.message)
    }
    call?.OK?.(response?.data)
    return response.data
}

const AddEquipment = async (
    params: EquipmentAddReq,
    call?: Callback<any, string>
) => {
    call?.setLoading?.(true)
    const response = await POST<Response>(URL.equipmentAdd, {
        ...params
    })
    call?.setLoading?.(false)
    if (response?.code !== 0) {
        call?.ERROR?.(response?.message)
        return Promise.reject(response?.message)
    }
    call?.OK?.(response?.data)
    return response?.data
}

const UpdateEquipment = async (
    id: string,
    params: EquipmentAddReq,
    call?: Callback<any, string>
) => {
    call?.setLoading?.(true)
    const response = await POST<Response>(URL.equipmentUpdate, {
        ...params,
        id
    })
    call?.setLoading?.(false)
    if (response?.code !== 0) {
        call?.ERROR?.(response?.message)
        return Promise.reject(response?.message)
    }
    call?.OK?.(response?.data)
    return response?.data
}

const GetEquipmentDetail = async (id: string, call?: Callback) => {
    if (!id) {
        call?.ERROR?.('id不能为空')
        return Promise.reject('id不能为空')
    }
    const res = await GET<Response<EquipmentListItem[]>>(URL.equipmentDetail, {
        id: id
    })
    if (res?.code !== 0) {
        call?.ERROR?.(res?.message)
        return Promise.reject(res?.message)
    }

    if (res.data.length >= 1) {
        return res.data[0]
    }
    call?.ERROR?.('未找到设备')
    return Promise.reject('未找到设备')
}

const DeleteEquipment = async (id: string) => {
    const response = await PostForm<Response>(URL.equipmentDelete, {
        id
    })
    if (response?.code !== 0) {
        return Promise.reject(response.message)
    }
    return response
}

export {
    GetEquipmentList,
    AddEquipment,
    UpdateEquipment,
    GetEquipmentDetail,
    DeleteEquipment
}
