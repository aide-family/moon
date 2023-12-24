import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Spin, Watermark } from 'antd'
import { FC, useContext, useEffect, useState } from 'react'
import { GlobalContext } from '@/context'
import roleOptions from '../options'
import { AES_Encrypt } from '@/utils/aes'
import roleApi from '@/apis/home/system/role'
import type { RoleListItem } from '@/apis/home/system/role/types'

export type EditModalProps = {
    open: boolean
    onClose: () => void
    id?: number
    onOk: () => void
}

const { roleCreate, roleDetail, roleUpdate } = roleApi
const { addFormItems, editFormItems } = roleOptions()

let timer: NodeJS.Timeout | null = null

const EditModal: FC<EditModalProps> = (props) => {
    const { user } = useContext(GlobalContext)
    const { open, onClose, id, onOk } = props
    const [form] = Form.useForm()
    const [loading, setLoading] = useState<boolean>(false)
    const [data, setData] = useState<RoleListItem | undefined>()

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
        form.validateFields().then((val) => {
            if (id) {
                // 编辑
                setLoading(true)
                roleUpdate({
                    id,
                    ...val,
                    password: AES_Encrypt(val.password)
                })
                    .then(onOk)
                    .finally(() => {
                        setLoading(false)
                    })
            } else {
                // 创建
                setLoading(true)
                roleCreate({
                    ...val,
                    password: AES_Encrypt(val.password)
                })
                    .then(onOk)
                    .finally(() => {
                        setLoading(false)
                    })
            }
        })
    }

    const fetchDetail = () => {
        if (!id) {
            return
        }
        roleDetail({ id }).then((res) => {
            setData(res.detail)
            form.setFieldsValue(res.detail)
        })
    }

    const Title = () => {
        return id ? '编辑角色' : '新增角色'
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

    const handleClose = () => {
        onClose()
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
            onCancel={handleClose}
            width="50vw"
            footer={<Footer />}
        >
            <Spin spinning={loading}>
                <Watermark content={user?.username} className="wh100">
                    <DataForm
                        form={form}
                        items={id ? editFormItems : addFormItems}
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
