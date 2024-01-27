/**分配权限 */
import { FC, useEffect, useState } from 'react'
import { Form, message, Modal, Select, Spin } from 'antd'

import roleApi from '@/apis/home/system/role'
import authApi from '@/apis/home/system/auth'
import { ApiAuthListReq } from '@/apis/home/system/auth/types'
import { DefaultOptionType } from 'antd/es/select'

const { roleDetail, roleRelateApi } = roleApi
const { authApiSelect } = authApi

export type DetailProps = {
    roleId: number
    open: boolean
    onClose: () => void
    onOk: () => void
}

const defaultSearchData = {
    page: {
        curr: 1,
        size: 10
    },
    keyword: ''
}

const AuthConfigModal: FC<DetailProps> = (props) => {
    const { open, onClose, onOk, roleId } = props
    const [form] = Form.useForm<{ roleIds: number[] }>()
    const [options, setOptions] = useState<DefaultOptionType[]>([])
    const [searchData, setSearchData] =
        useState<ApiAuthListReq>(defaultSearchData)

    const [fetchLoading, setFetchLoading] = useState<boolean>(false)
    const [submitLoading, setSubmitLoading] = useState<boolean>(false)

    const fetchUserDetail = async () => {
        if (!roleId) return
        const {
            detail: { apis }
        } = await roleDetail({ id: roleId })
        if (!apis) return
        const optionList = apis?.filter((item) => {
            return item.status === 1 && options.indexOf(item) === -1
        })
        setOptions([...options, ...optionList])
    }

    const handleGetApiSelect = () => {
        setFetchLoading(true)
        authApiSelect(searchData)
            .then(({ list, page: { total } }) => {
                if (!list || list.length === 0) return
                const optionList = list?.filter((item) => {
                    return item.status === 1 && options.indexOf(item) === -1
                })
                setOptions([...options, ...optionList])
                if (total > 0) {
                    setSearchData({
                        ...searchData,
                        page: { curr: searchData.page.curr + 1, size: 200 }
                    })
                }
            })
            .finally(() => {
                setFetchLoading(false)
            })
    }

    const handleAuthConfig = () => {
        setSubmitLoading(true)
        form.validateFields().then((data) => {
            roleRelateApi({ id: roleId, apiIds: data.roleIds })
                .then(() => {
                    onOk()
                    message.success('分配权限成功')
                })
                .finally(() => {
                    setSubmitLoading(false)
                })
        })
    }

    useEffect(() => {
        setOptions([])
        setSearchData(defaultSearchData)
        if (open) {
            fetchUserDetail()
            return
        }
    }, [open])

    useEffect(() => {
        handleGetApiSelect()
    }, [searchData])

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
                <Form form={form} layout="vertical">
                    <Form.Item label="权限" name="roleIds">
                        <Select
                            // loading={fetchLoading}
                            mode="multiple"
                            // style={{width: '100%', height: 300}}
                            style={{ width: '100%' }}
                            placeholder="请选择权限"
                            // value={selectedApi}
                            // onChange={handleChange}
                            options={options}
                            allowClear
                        />
                    </Form.Item>
                </Form>
            </Spin>
        </Modal>
    )
}

export default AuthConfigModal
