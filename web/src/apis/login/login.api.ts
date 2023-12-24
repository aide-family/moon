import { POST } from '../request'
import { LoginReq, LoginRes } from './types'

enum URL {
    LOGIN = '/api/v1/auth/login',
    LOGOUT = '/api/v1/auth/logout'
}

/**
 * 登录
 * @returns Promise<LoginRes>
 * @param data
 */
const login = (data: LoginReq): Promise<LoginRes> => {
    return POST<LoginRes>(URL.LOGIN, data)
}

export { login }
