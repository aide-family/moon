import type { UserListItem } from '../home/system/user/types'

export type LoginReq = {
    username: string
    password: string
    code: string
    captchaId: string
}

export type LoginRes = {
    token: string
    user: UserListItem
}

export type RefreshTokenResponse = {
    token: string
    user: UserListItem
}
