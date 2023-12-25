import {GET} from '../request'

export enum CaptchaType {
    CaptchaTypeUnknown,
    CaptchaTypeAudio,
    CaptchaTypeString,
    CaptchaTypeMath,
    CaptchaTypeChinese,
    CaptchaTypeDigit
}

export type CaptchaReq = {
    x?: number
    y?: number
    captchaType?: CaptchaType
}

export type CaptchaRes = {
    captcha: string
    captchaId: string
}

const captchaApi = '/api/v1/auth/captcha'

export const getCaptcha = (data: CaptchaReq): Promise<CaptchaRes> => {
    return GET(captchaApi, data)
}
