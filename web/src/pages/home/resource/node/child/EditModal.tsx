import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Spin, Watermark, message } from 'antd'
import { FC, useContext, useEffect, useState } from 'react'
import { editFormItems } from '../options'
import { NodeItemType } from '../type'
import { GlobalContext } from '@/context'
import { AddNode, GetNodeDetail, UpdateNode } from '@/apis/home/resource/node'

export type EditModalProps = {
    open: boolean
    onClose: () => void
    id?: string
    onOk: () => void
}

let timer: NodeJS.Timeout | null = null

const EditModal: FC<EditModalProps> = (props) => {
    const { user } = useContext(GlobalContext)
    const { open, onClose, id, onOk } = props
    const [form] = Form.useForm()
    const [loading, setLoading] = useState<boolean>(false)
    const [data, setData] = useState<NodeItemType | undefined>()

    const handleReset = () => {
        form.resetFields()
        form.setFieldsValue(data)
    }

    const handleCancel = () => {
        //  执行取消
        timer && clearTimeout(timer)
        onClose()
        setLoading(false)
    }

    const handleSubmit = () => {
        timer && clearTimeout(timer)
        timer = setTimeout(async () => {
            try {
                setLoading(true)
                if (id) {
                    // 编辑
                    await UpdateNode(id, form.getFieldsValue())
                } else {
                    // 创建
                    await AddNode(form.getFieldsValue())
                }
                onOk?.()
            } catch (e: any) {
                message.error(e)
            }

            setLoading(false)
        }, 500)
    }

    const fetchDetail = () => {
        if (!id) {
            return
        }
        GetNodeDetail(id || '', {
            ERROR: (msg) => {
                message.error(msg)
            },
            setLoading,
            OK: (res) => {
                setData(res)
                form.setFieldsValue(res)
            }
        })
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
