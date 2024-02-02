import DataForm from '@/components/Data/DataForm/DataForm'
import { Form, Modal } from 'antd'
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
                onCancel={handleClose}
                onOk={handleCommit}
                confirmLoading={loading}
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
