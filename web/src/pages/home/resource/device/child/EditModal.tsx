import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Spin, Watermark, message } from 'antd'
import { FC, useContext, useEffect, useState } from 'react'
import { editFormItems } from '../options'
import { GlobalContext } from '@/context'
import { DeviceItemType } from '../type'
import {
    AddEquipment,
    GetEquipmentDetail,
    UpdateEquipment
} from '@/apis/home/resource/device'

export type EditModalProps = {
    open: boolean
    onClose: () => void
    id?: string
    onOk?: () => void
}

let timer: NodeJS.Timeout | null = null

const EditModal: FC<EditModalProps> = (props) => {
    const { user } = useContext(GlobalContext)
    const { open, onClose, onOk, id } = props
    const [data, setData] = useState<DeviceItemType | undefined>()
    const [loading, setLoading] = useState<boolean>(false)
    const [form] = Form.useForm()

    const fetchDetail = () => {
        if (!id) {
            return
        }
        setLoading(true)
        GetEquipmentDetail(id || '', {
            ERROR: (msg) => {
                message.error(msg)
            }
        })
            .then((res) => {
                setData(res)
                form.setFieldsValue(res)
            })
            .finally(() => setLoading(false))
    }

    const handleReset = () => {
        form.resetFields()
        form.setFieldsValue(data)
    }

    const handleSubmit = () => {
        timer && clearTimeout(timer)
        timer = setTimeout(async () => {
            try {
                setLoading(true)
                if (id) {
                    // 编辑
                    await UpdateEquipment(id, form.getFieldsValue())
                } else {
                    // 创建
                    await AddEquipment(form.getFieldsValue())
                }
                onOk?.()
            } catch (e: any) {
                message.error(e)
            }

            setLoading(false)
        }, 500)
    }

    const handleCancel = () => {
        //  执行取消
        timer && clearTimeout(timer)
        onClose()
        setLoading(false)
    }

    const Title = () => {
        return id ? '编辑设备' : '新增设备'
    }

    const Footer = () => {
        return (
            <>
                <Button type="dashed" onClick={handleCancel}>
                    取消
                </Button>
                <Button type="default" onClick={handleReset} loading={loading}>
                    恢复
                </Button>
                <Button type="primary" onClick={handleSubmit} loading={loading}>
                    确定
                </Button>
            </>
        )
    }

    useEffect(() => {
        if (open) {
            fetchDetail()
        } else {
            setData(undefined)
            form.resetFields()
        }
    }, [open])

    return (
        <Modal
            title={<Title />}
            open={open}
            onCancel={onClose}
            width="50vw"
            footer={<Footer />}
        >
            <Spin spinning={loading}>
                <Watermark content={user?.user_name} className="wh100">
                    <DataForm
                        form={form}
                        items={editFormItems}
                        formProps={{
                            layout: 'vertical'
                        }}
                    />
                </Watermark>
            </Spin>
        </Modal>
    )
}

export default EditModal
