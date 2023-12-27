import { FC } from 'react'
import { labelsType } from './StrategyForm'
import { Form, FormInstance, Modal } from 'antd'
import DataForm, { DataFormItem } from '@/components/Data/DataForm/DataForm'

export interface AddLabelModalProps {
    open: boolean
    onCancel: () => void
    onOk: (data: labelsType) => void
}

export interface AddLabelFormProps {
    form: FormInstance
}

const items: DataFormItem[] = [
    {
        label: '标签Label',
        name: 'label',
        rules: [
            {
                required: true,
                message: '请输入标签名称'
            }
        ]
    },
    {
        label: '标签Key',
        name: 'name',
        rules: [
            {
                required: true,
                message: '请输入标签Key'
            }
        ]
    }
]

const AddLabelForm: FC<AddLabelFormProps> = (props) => {
    const { form } = props
    return (
        <div>
            <DataForm
                formProps={{ layout: 'vertical' }}
                form={form}
                items={items}
            />
        </div>
    )
}

const AddLabelModal: FC<AddLabelModalProps> = (props) => {
    const { open, onCancel, onOk } = props
    const [form] = Form.useForm()

    const handleOnOK = () => {
        form.validateFields().then((values) => {
            onOk(values)
            form.resetFields()
        })
    }
    return (
        <Modal
            title="添加标签"
            open={open}
            onCancel={onCancel}
            onOk={handleOnOK}
            width={600}
        >
            <AddLabelForm form={form} />
        </Modal>
    )
}

export default AddLabelModal
