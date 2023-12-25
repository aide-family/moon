import DataForm from '@/components/Data/DataForm/DataForm'
import {Button, Form, Modal, Spin, Watermark} from 'antd'
import {FC, useContext, useEffect, useState} from 'react'
import {GlobalContext} from '@/context'
import dictOptions from '../options'
import {AES_Encrypt} from '@/utils/aes'
import dict from '@/apis/home/system/dict'
import type {DictListItem} from '@/apis/home/system/dict/types'

const {dictCreate, dictUpdate, dictDetail} = dict
const {addFormItems, editFormItems} = dictOptions()

// TODO 待测试字典编辑功能

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
    const [data, setData] = useState<DictListItem | undefined>()

    /** 重置 */
    const handleReset = () => {
        form.resetFields()
        form.setFieldsValue(data)
    }

    /** 执行取消 */
    const handleCancel = () => {
        timer && clearTimeout(timer)
        onClose()
        setLoading(false)
    }

    /** 关闭 */
    const handleClose = () => {
        onClose()
    }

    /** 编辑 */
    const onEdit = (val: any) => {
        setLoading(true)
        dictUpdate({
            id,
            ...val,
            password: AES_Encrypt(val.password)
        })
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    /** 创建 */
    const onCreate = (val: any) => {
        setLoading(true)
        dictCreate({...val, password: AES_Encrypt(val.password)})
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    /** 提交 */
    const handleSubmit = () => {
        form.validateFields().then((val) => {
            if (id) {
                onEdit(val)
                return
            }
            onCreate(val)
        })
    }

    const fetchDetail = () => {
        if (!id) {
            return
        }
        dictDetail({id}).then((res) => {
            setData(res.promDict)
            form.setFieldsValue(res.promDict)
        })
    }

    const Title = () => {
        return id ? '编辑字典' : '新增字典'
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
