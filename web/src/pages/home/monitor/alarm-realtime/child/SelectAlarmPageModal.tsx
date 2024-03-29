import dictApi from '@/apis/home/system/dict'
import { BindMyAlarmPagesRequest } from '@/apis/home/system/dict/types'
import { Category } from '@/apis/types'
import FetchSelect from '@/components/Data/FetchSelect'
import { SettingOutlined } from '@ant-design/icons'
import { Button, Form, Modal, Tag } from 'antd'
import { DefaultOptionType } from 'antd/es/select'
import React, { useEffect, useState } from 'react'

export interface SelectAalrmPageModalProps {
    refresh?: () => void
}

export const SelectAalrmPageModal: React.FC<SelectAalrmPageModalProps> = (
    props
) => {
    const { refresh } = props

    const [form] = Form.useForm<BindMyAlarmPagesRequest>()

    const [open, setOpen] = useState(false)

    const handleClose = () => setOpen(false)

    const handleOpen = () => setOpen(true)

    const handleOnConfirm = () => {
        form.validateFields().then((formVal) => {
            dictApi.myAlarmPageConfig(formVal).then(() => {
                handleClose()
                refresh?.()
            })
        })
    }

    const handleGetMyAlarmPages = () => {
        dictApi.myAlarmPageList().then(({ list }) => {
            form.setFieldsValue({
                alarmIds: list.map(({ id }) => id)
            })
        })
    }

    const handleGetAlarmPage = (keyword: string) => {
        return dictApi
            .dictSelect({
                page: { curr: 1, size: 10 },
                keyword,
                category: Category.CATEGORY_ALARM_PAGE
            })
            .then(({ list }): DefaultOptionType[] => {
                if (!list || list.length === 0) return []
                return list.map(({ value, label, color }) => ({
                    label: <Tag color={color}>{label}</Tag>,
                    value: value
                }))
            })
    }

    useEffect(() => {
        if (!open) return
        handleGetMyAlarmPages()
    }, [open])

    return (
        <>
            <Modal
                title="告警页面配置"
                open={open}
                onOk={handleOnConfirm}
                onCancel={handleClose}
            >
                <Form form={form} layout="vertical">
                    <Form.Item
                        name="alarmIds"
                        label="告警页面"
                        rules={[
                            {
                                required: true,
                                message: '请选择告警页面'
                            }
                        ]}
                        tooltip="绑定你的告警页面，最多同时展示10个页面"
                    >
                        <FetchSelect
                            defaultOptions={[]}
                            handleFetch={handleGetAlarmPage}
                            selectProps={{
                                mode: 'multiple',
                                placeholder: '请选择你要关注的报警页面',
                                maxCount: 10
                            }}
                        />
                    </Form.Item>
                </Form>
            </Modal>
            <Button
                type="primary"
                icon={<SettingOutlined />}
                onClick={handleOpen}
            />
        </>
    )
}
