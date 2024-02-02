import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal } from 'antd'
import React, { useState } from 'react'
import { addDashboardOptions } from '../options'
import { CreateDashboardRequest } from '@/apis/home/dashboard/types'
import dashboardApi from '@/apis/home/dashboard'

export interface AddDashboardButtonProps {
    refresh?: () => void
}

export const AddDashboardButton: React.FC<AddDashboardButtonProps> = (
    props
) => {
    const { refresh } = props

    const [form] = Form.useForm<CreateDashboardRequest>()

    const [open, setOpen] = useState(false)
    const [loading, setLoading] = useState<boolean>(false)

    const handleOpen = () => setOpen(true)

    const handleClose = () => setOpen(false)

    const handleCommit = () => {
        form.validateFields().then((values) => {
            setLoading(true)
            dashboardApi
                .createDashboard(values)
                .then(() => {
                    handleClose()
                    refresh?.()
                })
                .finally(() => setLoading(false))
        })
    }

    return (
        <>
            <Modal
                title="添加大盘"
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
            <Button type="primary" onClick={handleOpen}>
                添加大盘
            </Button>
        </>
    )
}
