import DataForm from '@/components/Data/DataForm/DataForm'
import { Form, Modal } from 'antd'
import React, { useEffect, useState } from 'react'
import { alarmPageDataFormItems, alarmPageDataFormType } from '../options'
import alarmPageApi from '@/apis/home/monitor/alarm-page'
import { AlarmPageItem } from '@/apis/home/monitor/alarm-page/types'

export interface EditAlarmPageModalProps {
    open: boolean
    onCancel: () => void
    id?: number
    onOk: () => void
}

const EditAlarmPageModal: React.FC<EditAlarmPageModalProps> = (props) => {
    const { open, onCancel, id, onOk } = props
    const [form] = Form.useForm<alarmPageDataFormType>()
    const [loading, setLoading] = useState(false)
    const [alarmpageDetail, setAlarmpageDetail] = useState<
        AlarmPageItem | undefined
    >()

    const Title = () => {
        return id ? <div>编辑告警页面</div> : <div>新增告警页面</div>
    }

    const handleOnCancel = () => {
        onCancel()
    }

    const handleGetAlarmPageDetail = () => {
        if (!id) {
            return
        }
        return alarmPageApi.getAlarmPageDetail({ id }).then((res) => {
            setAlarmpageDetail(res.alarmPage)
        })
    }

    const handleEditAlarmPage = (value: alarmPageDataFormType) => {
        if (!id) {
            return Promise.reject(new Error('id is undefined'))
        }
        return alarmPageApi.updateAlarmPage({ id, ...value })
    }

    const handleAddAlarmPage = (value: alarmPageDataFormType) => {
        return alarmPageApi.createAlarmPage({ ...value })
    }

    const handleNewAlarmPage = (values: alarmPageDataFormType) => {
        if (id) {
            return handleEditAlarmPage(values)
        }

        return handleAddAlarmPage(values)
    }

    const handleOnOk = () => {
        form.resetFields()
        onOk()
    }

    const handleOk = () => {
        setLoading(true)
        form.validateFields()
            .then((values) => {
                return handleNewAlarmPage(values)
            })
            .then(handleOnOk)
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        if (!alarmpageDetail || !form) {
            return
        }
        form.setFieldsValue(alarmpageDetail)
    }, [alarmpageDetail])

    useEffect(() => {
        handleGetAlarmPageDetail()
    }, [id])

    return (
        <Modal
            title={<Title />}
            open={open}
            onCancel={handleOnCancel}
            onOk={handleOk}
            confirmLoading={loading}
        >
            <DataForm
                form={form}
                formProps={{ layout: 'vertical' }}
                items={alarmPageDataFormItems}
            />
        </Modal>
    )
}

export default EditAlarmPageModal
