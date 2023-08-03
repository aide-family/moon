import {Message} from '@arco-design/web-react'
import axios from 'axios'

export type Resp = {
    code: number
    data: any
    message: string
    metadata?: { [key: string]: string }
}

export type PageRequest = {
    current: number
    size: number
}

export type PageReply = {
    current: number
    total: string
    size: number
}

export const REACT_APP_MASTER_API = process.env.REACT_APP_MASTER_API

const request = axios.create({
    baseURL: REACT_APP_MASTER_API,
    timeout: 10000,
})

request.interceptors.request.use((config) => {
    return config
})

request.interceptors.response.use(
    (response) => {
        return response?.data
    },
    (error) => {
        console.log(error);
        if (error?.response?.status !== 200) {
            let msg = error?.response?.data?.message || '系统异常，请稍后再试'
            Message.error(msg)
            return Promise.reject(msg)
        }

        let msg = error?.response?.data?.message || '系统异常，请稍后再试'
        Message.error(msg)

        return Promise.reject(error?.response?.data || error)
    }
)

const Get = async <T>(url: string, params?: any, headers?: any) => {
    const data = await request({
        url,
        method: 'get',
        params,
        headers,
    })
    return data as T
}

const Post = <T>(url: string, body?: any, headers?: any) => {
    return request({
        url,
        method: 'post',
        data: body,
        headers,
    }).then((res) => res as T)
}

const Put = <T>(url: string, data?: any, headers?: any) => {
    return request({
        url,
        method: 'put',
        data,
        headers,
    }) as Promise<T>
}

const Delete = <T>(url: string, data?: any, headers?: any) => {
    return request({
        url,
        method: 'delete',
        data,
        headers,
    }) as Promise<T>
}

export {request, Get, Post, Put, Delete}
