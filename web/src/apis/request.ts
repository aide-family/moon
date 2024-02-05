import axios, { AxiosError, AxiosPromise, AxiosRequestConfig } from 'axios'

import type { Response, Map } from '@/apis/types.ts'
import { notification } from 'antd'

type ErrorRepose = {
    code: number
    message: string
    metadata: Map<string>
    reason: string
}
const localhost = 'http://localhost:5173'
const localhostDocker = 'http://localhost:8000'
const host =
    window.location.origin || process.env.REACT_APP_ASSET_API || localhostDocker

const hostMap: { [key: string]: string } = {
    [localhost]: 'http://localhost:8000',
    [localhostDocker]: 'http://localhost:8001'
}

const request = axios.create({
    baseURL: hostMap[host] || host,
    timeout: 10000
})

const info = (msg?: AxiosError<ErrorRepose>) => {
    if (!msg || !msg.response) return
    const {
        code,
        message = '系统错误',
        reason = 'SYSTEM_ERROR'
    } = msg?.response?.data

    if (code === 401) {
        setTimeout(() => {
            window.location.href = '/#/login'
        }, 1000)
        return
    }
    notification.open({
        message: reason,
        description: message,
        type: 'error',
        duration: 3,
        role: 'alert'
    })
}

request.interceptors.request.use((config) => {
    const token = localStorage.getItem('token') || ''
    config.headers['Authorization'] = `Bearer ${token}`
    return config
})

request.interceptors.response.use(
    (response): AxiosPromise<Response> => {
        return response.data
    },
    (error: AxiosError<any>) => {
        info(error)
        return Promise.reject(error)
    }
)

const POST = <T = Response>(
    url: string,
    data?: Map,
    config?: AxiosRequestConfig<any>
): Promise<T> => {
    return request.post(url, data, config) as Promise<T>
}

const PostForm = <T = Response>(
    url: string,
    data: Map,
    config?: AxiosRequestConfig<Map>
): Promise<T> => {
    return request.post(url, data, {
        ...config,
        headers: {
            ...config?.headers,
            'Content-Type': 'application/x-www-form-urlencoded'
        }
    })
}

const GET = <T = Response>(
    url: string,
    data?: Map,
    config?: AxiosRequestConfig<Map>
): Promise<T> => {
    return request.get(url, {
        ...config,
        params: data
    })
}

const PUT = <T = Response>(
    url: string,
    data?: Map,
    config?: AxiosRequestConfig<Map>
): Promise<T> => {
    return request.put(url, data, config)
}

const DELETE = <T = Response>(
    url: string,
    data?: Map,
    config?: AxiosRequestConfig<Map>
): Promise<T> => {
    return request.delete(url, {
        ...config,
        params: data
    })
}

export { request, POST, PostForm, GET, PUT, DELETE }
