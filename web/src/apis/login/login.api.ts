import { POST } from '../request'
import { LoginReq, LoginRes, RefreshTokenResponse } from './types'

enum URL {
    LOGIN = '/api/v1/auth/login',
    LOGOUT = '/api/v1/auth/logout',
    REFRESH_TOKEN = '/api/v1/auth/refresh/token'
}

/**
 * 登录
 * @returns Promise<LoginRes>
 * @param data
 */
const login = (data: LoginReq): Promise<LoginRes> => {
    return POST<LoginRes>(URL.LOGIN, data)
}

const logout = (): Promise<LoginRes> => {
    return POST(URL.LOGOUT)
}

const refreshToken = (roleId?: number): Promise<LoginRes> => {
    return POST<RefreshTokenResponse>(URL.REFRESH_TOKEN, {
        roleId: roleId || 0
    })
}

export { login, logout, refreshToken }
