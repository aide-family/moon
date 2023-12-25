import DataForm from '@/components/Data/DataForm/DataForm'
import {Button, Form, Modal, Spin, Watermark} from 'antd'
import {FC, useContext, useEffect, useState} from 'react'
import {GlobalContext} from '@/context'
import userOptions from '../options'
import userApi from '@/apis/home/system/user'
import {AES_Encrypt} from '@/utils/aes'
import type {UserListItem} from '@/apis/home/system/user/types'

const {userDetail, userCreate} = userApi
const {addFormItems, editFormItems} = userOptions()

export type EditModalProps = {
    open: boolean
    onClose: () => void
    id?: number
    onOk: () => void
}

let timer: NodeJS.Timeout | null = null

const EditModal: FC<EditModalProps> = (props) => {
    const {user} = useContext(GlobalContext)
    const {open, onClose, id, onOk} = props
    const [form] = Form.useForm()
    const [loading, setLoading] = useState<boolean>(false)
    const [data, setData] = useState<UserListItem | undefined>()

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
                console.log('编辑', val)
            } else {
                // 创建
                console.log('创建', val)
                setLoading(true)
                userCreate({
                    ...val,
                    password: AES_Encrypt(val.password)
                })
                    .then(() => {
                        onOk?.()
                    })
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
        userDetail({id}).then((res) => {
            setData(res.detail)
            form.setFieldsValue(res)
        })
    }

    const Title = () => {
        return id ? '编辑用户' : '新增用户'
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
            title={<Title/>}
            open={open}
            onCancel={handleClose}
            width="50vw"
            footer={<Footer/>}
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
