import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal } from 'antd'
import React, { useState } from 'react'
import { addChartOptions } from '../options'
import { CreateChartRequest } from '@/apis/home/dashboard/types'
import dashboardApi from '@/apis/home/dashboard'

export interface AddDashboardButtonProps {}

export const AddChartButton: React.FC<AddDashboardButtonProps> = (props) => {
    const {} = props

    const [form] = Form.useForm<CreateChartRequest>()

    const [open, setOpen] = useState(false)
    const [loading, setLoading] = useState<boolean>(false)

    const handleOpen = () => {
        setOpen(true)
        form.resetFields()
    }

    const handleClose = () => {
        form.resetFields()
        setOpen(false)
    }

    const handleCommit = () => {
        form.validateFields().then((values) => {
            setLoading(true)
            dashboardApi
                .createChart(values)
                .then(handleClose)
                .finally(() => setLoading(false))
        })
    }

    return (
        <>
            <Modal
                title="添加图表"
                open={open}
                onCancel={handleClose}
                onOk={handleCommit}
                confirmLoading={loading}
                width='60%'
            >
                <DataForm
                    items={addChartOptions}
                    form={form}
                    formProps={{ layout: 'vertical' }}
                />
            </Modal>
            <Button type="primary" onClick={handleOpen}>
                添加图表
            </Button>
        </>
    )
}
