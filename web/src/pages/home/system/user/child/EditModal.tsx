import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Spin, Watermark } from 'antd'
import { FC, useContext, useEffect, useState } from 'react'
import { GlobalContext } from '@/context'
import userOptions from '../options'
import userApi from '@/apis/home/system/user'
import { AES_Encrypt } from '@/utils/aes'
import type {
    UserCreateParams,
    UserListItem,
    UserUpdateParams
} from '@/apis/home/system/user/types'

const { userDetail, userCreate, userUpdate } = userApi
const { addFormItems, editFormItems } = userOptions()

export type EditModalProps = {
    open: boolean
    onClose: () => void
    id?: number
    onOk: () => void
}

let timer: NodeJS.Timeout | null = null

const EditUserModal: FC<EditModalProps> = (props) => {
    const { user } = useContext(GlobalContext)
    const { open, onClose, id, onOk } = props
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

    const createUser = (userItem: UserCreateParams) => {
        setLoading(true)
        userCreate(userItem)
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const updateUser = (userItem: UserUpdateParams) => {
        setLoading(true)
        userUpdate(userItem)
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const handleSubmit = () => {
        form.validateFields().then((val) => {
            if (id) {
                updateUser({ ...val, id })
            } else {
                createUser({
                    ...val,
                    password: AES_Encrypt(val.password)
                })
            }
        })
    }

    const fetchDetail = () => {
        if (!id) {
            return
        }
        userDetail({ id }).then((res) => {
            const { detail } = res
            setData(detail)
            form.setFieldsValue({
                ...detail,
                roleIds: detail.roles?.map((item) => item.value)
            })
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

export default EditUserModal
