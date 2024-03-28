import { POST } from '@/apis/request'
import {
    CreateTemplateRequest,
    GetTemplateReply,
    ListTemplateReply,
    ListTemplateRequest,
    UpdateTemplateRequest
} from './types'

enum URL {
    GET_TEMPLATE = '/api/v1/notify/template/get',
    LIST_TEMPLATE = '/api/v1/notify/template/list',
    CREATE_TEMPLATE = '/api/v1/notify/template/create',
    UPDATE_TEMPLATE = '/api/v1/notify/template/update',
    DELETE_TEMPLATE = '/api/v1/notify/template/delete'
}

/**
 * 获取模板
 * @param templateId 模板id
 * @returns 模板详情
 */
const getTemplate = (templateId: number) => {
    return POST<GetTemplateReply>(URL.GET_TEMPLATE, { id: templateId })
}

/**
 * 获取模板列表
 * @param params 模板列表请求参数
 * @returns 模板列表
 */
const getTemplateList = (params: ListTemplateRequest) => {
    return POST<ListTemplateReply>(URL.LIST_TEMPLATE, params)
}

/**
 * 创建模板
 * @param params 模板创建请求参数
 * @returns 模板详情
 */
const createTemplate = (params: CreateTemplateRequest) => {
    return POST<{}>(URL.CREATE_TEMPLATE, params)
}

/**
 * 更新模板
 * @param params 模板更新请求参数
 * @returns 模板详情
 */
const updateTemplate = (params: UpdateTemplateRequest) => {
    return POST<{}>(URL.UPDATE_TEMPLATE, params)
}

/**
 * 删除模板
 * @param id 模板id
 * @returns 模板详情
 */
const deleteTemplate = (itemplateId: number) => {
    return POST<{}>(URL.DELETE_TEMPLATE, { id: itemplateId })
}

export const templateApi = {
    getTemplate,
    getTemplateList,
    createTemplate,
    updateTemplate,
    deleteTemplate
}
