import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Space } from 'antd'
import React, { useEffect, useState } from 'react'
import { addDashboardOptions } from '../options'
import { UpdateDashboardRequest } from '@/apis/home/dashboard/types'
import dashboardApi from '@/apis/home/dashboard'

export interface ConfigDashboardChartButtonProps {
    dashboardId?: number
    open: boolean
    onOk?: () => void
    onCancel?: () => void
}

export const ConfigDashboardChartModal: React.FC<
    ConfigDashboardChartButtonProps
> = (props) => {
    const { onOk, onCancel, dashboardId, open } = props

    const [form] = Form.useForm<UpdateDashboardRequest>()

    const [loading, setLoading] = useState<boolean>(false)

    const handleGetDashboardDetail = () => {
        if (!dashboardId) {
            return
        }
        dashboardApi.getDashboardDetail(dashboardId).then(({ detail }) => {
            if (!detail) {
                return
            }
            form.setFieldsValue({
                ...detail,
                chartIds: detail?.charts?.map((item) => item.id)
            })
        })
    }

    const handleClose = () => {
        form.resetFields()
        onCancel?.()
    }

    const handleOnOk = () => {
        form.resetFields()
        onOk?.()
    }

    const handleCommit = () => {
        if (!dashboardId) {
            return
        }
        form.validateFields().then((values) => {
            setLoading(true)
            dashboardApi
                .updateDashboard({ ...values, id: dashboardId })
                .then(() => {
                    handleOnOk()
                })
                .finally(() => setLoading(false))
        })
    }

    const handleDelete = () => {
        if (!dashboardId) {
            return
        }
        setLoading(true)
        dashboardApi
            .deleteDashboard(dashboardId)
            .then(() => {
                handleOnOk()
            })
            .finally(() => setLoading(false))
    }

    useEffect(() => {
        if (!open) {
            return
        }
        handleGetDashboardDetail()
    }, [open])

    return (
        <>
            <Modal
                title="配置大盘"
                open={open}
                footer={
                    <Space size={8}>
                        <Button
                            loading={loading}
                            danger
                            type="primary"
                            onClick={handleDelete}
                        >
                            删除大盘
                        </Button>
                        <Button onClick={handleClose}>取消</Button>
                        <Button
                            loading={loading}
                            type="primary"
                            onClick={handleCommit}
                        >
                            确定
                        </Button>
                    </Space>
                }
            >
                <DataForm
                    items={addDashboardOptions}
                    form={form}
                    formProps={{ layout: 'vertical' }}
                />
            </Modal>
        </>
    )
}
