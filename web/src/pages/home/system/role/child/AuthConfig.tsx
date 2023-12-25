/**分配权限 */
import {FC, useEffect, useState} from 'react'
import {message, Modal, Select, Spin} from 'antd'

import roleApi from '@/apis/home/system/role'
import authApi from '@/apis/home/system/auth'
import {ApiAuthListReq, ApiAuthSelectItem} from '@/apis/home/system/auth/types'

const {roleDetail, roleRelateApi} = roleApi
const {authApiSelect} = authApi

export type DetailProps = {
    roleId: number
    open: boolean
    onClose: () => void
    onOk: () => void
}
/** 权限配置组件
 * @param props type: DetailProps
 * @type roleId : number   // 用户id
 * @type open : boolean    // 是否显示
 * @type onClose : () => void // 关闭回调
 *
 */
const AuthConfigModal: FC<DetailProps> = (props) => {
    const {open, onClose, onOk, roleId} = props

    const defaultSearchData = {
        page: {
            curr: 1,
            size: 10
        },
        keyword: ''
    }
    const [options, setOptions] = useState<ApiAuthSelectItem[]>([])
    const [searchData, setSearchData] =
        useState<ApiAuthListReq>(defaultSearchData)

    const [selectedApi, setSelectedApi] = useState<number[]>([])
    const [fetchLoading, setFetchLoading] = useState<boolean>(false)
    const [submitLoading, setSubmitLoading] = useState<boolean>(false)

    const fetchUserDetail = async () => {
        const res = await roleDetail({id: roleId})
        setSelectedApi(res.detail.apis?.map((item) => item.value) || [])
        setOptions([...new Set([...options, ...(res.detail.apis || [])])])
    }
    useEffect(() => {
        setOptions([])
        setSearchData(defaultSearchData)
        if (open) {
            fetchUserDetail()
            return
        }
        setSelectedApi([])
    }, [open, roleId])

    useEffect(() => {
        if (open) {
            setFetchLoading(true)
            authApiSelect(searchData)
                .then((res) => {
                    console.log(res)
                    //追加 赋值给options 并去重
                    setOptions([...new Set([...options, ...res.list])])
                })
                .finally(() => {
                    setFetchLoading(false)
                })
        }
    }, [searchData, open, roleId])

    const handleChange = (value: number[]) => {
        setSelectedApi(value)
    }
    const handleScroll = (e: any) => {
        const {scrollTop, clientHeight, scrollHeight} = e.target
        // 滚动到底部时加载更多数据
        if (scrollHeight - scrollTop === clientHeight) {
            // setPage((prevPage) => prevPage + 1)
            setSearchData({
                ...searchData,
                page: {curr: searchData.page.curr + 1, size: 10}
            })
            console.log('滚动到底部')
        }
    }
    const handleAuthConfig = () => {
        setSubmitLoading(true)
        console.log('分配权限')
        roleRelateApi({id: roleId, apiIds: selectedApi})
            .then(() => {
                onOk()
                message.success('分配权限成功')
            })
            .finally(() => {
                setSubmitLoading(false)
            })
    }

    return (
        <Modal
            open={open}
            onCancel={onClose}
            centered
            keyboard={false}
            title="分配权限"
            onOk={handleAuthConfig}
            confirmLoading={submitLoading}
        >
            <Spin spinning={fetchLoading} tip="加载中...">
                <Select
                    loading={fetchLoading}
                    mode="multiple"
                    style={{width: '100%', height: 300}}
                    placeholder="请选择权限"
                    value={selectedApi}
                    onChange={handleChange}
                    options={options}
                    onPopupScroll={handleScroll}
                />
            </Spin>
        </Modal>
    )
}

export default AuthConfigModal
