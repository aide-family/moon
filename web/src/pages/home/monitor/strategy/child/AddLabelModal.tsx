import { FC } from 'react'
import { labelsType } from './StrategyForm'
import { Form, FormInstance, Modal } from 'antd'
import DataForm, { DataFormItem } from '@/components/Data/DataForm/DataForm'

export interface AddLabelModalProps {
    open: boolean
    onCancel: () => void
    onOk: (data: labelsType) => void
    title: string
}

export interface AddLabelFormProps {
    form: FormInstance
    title: string
}

const items = (title: string): DataFormItem[] => [
    {
        label: `新${title}`,
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
    const { form, title } = props
    return (
        <div>
            <DataForm
                formProps={{ layout: 'vertical' }}
                form={form}
                items={items(title)}
            />
        </div>
    )
}

const AddLabelModal: FC<AddLabelModalProps> = (props) => {
    const { open, onCancel, onOk, title } = props
    const [form] = Form.useForm()

    const handleOnOK = () => {
        form.validateFields().then(({ name }) => {
            onOk({
                name: name,
                label: name
            })
            form.resetFields()
        })
    }
    return (
        <Modal
            title={`添加${title}`}
            open={open}
            onCancel={onCancel}
            onOk={handleOnOK}
            width={600}
        >
            <AddLabelForm title={title} form={form} />
        </Modal>
    )
}

export default AddLabelModal
