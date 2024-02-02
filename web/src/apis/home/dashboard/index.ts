import { POST } from '@/apis/request'
import {
    CreateChartRequest,
    CreateDashboardRequest,
    GetChartResponse,
    GetDashboardResponse,
    ListChartRequest,
    ListChartResponse,
    ListDashboardRequest,
    ListDashboardResponse,
    SelectDashboardOptionRequest,
    SelectDashboardOptionResponse,
    UpdateChartRequest,
    UpdateDashboardRequest
} from './types'

enum URL {
    /** 创建新图表 */
    CREATE_CHART = '/api/v1/dashboard/chart/create',
    /** 删除图表 */
    DELETE_CHART = '/api/v1/dashboard/chart/delete',
    /** 更新图表 */
    UPDATE_CHART = '/api/v1/dashboard/chart/update',
    /** 获取图表列表 */
    GET_CHART_LIST = '/api/v1/dashboard/chart/list',
    /** 获取图表详情 */
    GET_CHART_DETAIL = '/api/v1/dashboard/chart/detail',
    /* --------------------分割线-------------------------- */
    /** 创建新的大盘 */
    CREATE_DASHBOARD = '/api/v1/dashboard/create',
    /** 删除大盘 */
    DELETE_DASHBOARD = '/api/v1/dashboard/delete',
    /** 更新大盘 */
    UPDATE_DASHBOARD = '/api/v1/dashboard/update',
    /** 获取大盘列表 */
    GET_DASHBOARD_LIST = '/api/v1/dashboard/list',
    /** 获取大盘详情 */
    GET_DASHBOARD_DETAIL = '/api/v1/dashboard/detail',
    /** 大盘下拉列表 */
    GET_DASHBOARD_SELECT = '/api/v1/dashboard/select'
    /* --------------------分割线-------------------------- */
}

/**
 * 创建图表
 * @param data
 * @returns
 */
const createChart = (data: CreateChartRequest) => {
    return POST<{ id: number }>(URL.CREATE_CHART, data)
}

/**
 * 删除图表
 * @param id
 * @returns
 */
const deleteChart = (id: number) => {
    return POST<{ id: number }>(URL.DELETE_CHART, { id })
}

/**
 * 更新图表
 * @param data
 * @returns
 */
const updateChart = (data: UpdateChartRequest) => {
    return POST<{ id: number }>(URL.UPDATE_CHART, data)
}

/**
 * 获取图表列表
 * @param data
 * @returns
 */
const getChartList = (data: ListChartRequest) => {
    return POST<ListChartResponse>(URL.GET_CHART_LIST, data)
}

/**
 * 获取图表详情
 * @param id
 * @returns
 */
const getChartDetail = (id: number) => {
    return POST<GetChartResponse>(URL.GET_CHART_DETAIL, { id })
}

/**
 * 创建大盘
 * @param data
 * @returns
 */
const createDashboard = (data: CreateDashboardRequest) => {
    return POST<{ id: number }>(URL.CREATE_DASHBOARD, data)
}

/**
 * 删除大盘
 * @param id
 * @returns
 */
const deleteDashboard = (id: number) => {
    return POST<{ id: number }>(URL.DELETE_DASHBOARD, { id })
}

/**
 * 更新大盘
 * @param data
 * @returns
 */
const updateDashboard = (data: UpdateDashboardRequest) => {
    return POST<{ id: number }>(URL.UPDATE_DASHBOARD, data)
}

/**
 * 获取大盘列表
 * @param data
 * @returns
 */
const getDashboardList = (data: ListDashboardRequest) => {
    return POST<ListDashboardResponse>(URL.GET_DASHBOARD_LIST, data)
}

/**
 * 获取大盘详情
 * @param id
 * @returns
 */
const getDashboardDetail = (id: number) => {
    return POST<GetDashboardResponse>(URL.GET_DASHBOARD_DETAIL, { id })
}

/**
 * 获取大盘下拉列表
 * @param data
 * @returns
 */
const getDashboardSelect = (data: SelectDashboardOptionRequest) => {
    return POST<SelectDashboardOptionResponse>(URL.GET_DASHBOARD_SELECT, data)
}

const dashboardApi = {
    createChart,
    deleteChart,
    updateChart,
    getChartList,
    getChartDetail,
    createDashboard,
    deleteDashboard,
    updateDashboard,
    getDashboardList,
    getDashboardDetail,
    getDashboardSelect
}

export default dashboardApi
