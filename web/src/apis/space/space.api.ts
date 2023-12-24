import { GET } from '../request'
import { Response } from '../types'
const apiUrlMap: Record<string, string> = {
    getSpace: '/space/instance/v1/select/user/space'
}

export type GetSpaceRes = {
    id: string
    name: string
    logo?: string
    is_team: number
}

/**
 * 获取用户空间
 * @returns Promise<GetSpaceRes>
 */
export const GetSpace = () => {
    return GET<Response<any[]>>(
        apiUrlMap.getSpace,
        {},
        {
            baseURL: process.env.REACT_APP_SECURITY_API
        }
    ).then((res) => {
        // 异常
        if (res?.code !== 0) {
            return Promise.reject(res?.message)
        }

        return res.data
    })
}
