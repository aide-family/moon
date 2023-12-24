import type { DataFormItem } from '@/components/Data/DataForm/DataForm'
import { UserOutlined, LockOutlined } from '@ant-design/icons'

export const formItems: DataFormItem[] = [
    {
        name: 'username',
        label: '',
        dataProps: {
            type: 'input',
            parentProps: {
                placeholder: '请输入用户名',
                size: 'large',
                prefix: <UserOutlined />
            }
        },
        rules: [
            {
                required: true,
                message: '请输入用户名'
            }
        ]
    },
    {
        name: 'password',
        label: '',
        dataProps: {
            type: 'password',
            parentProps: {
                placeholder: '请输入密码',
                size: 'large',
                prefix: <LockOutlined />
            }
        },
        rules: [
            {
                required: true,
                message: '请输入密码'
            }
        ]
    },
    {
        name: 'code',
        label: '',
        rules: [
            {
                required: true,
                message: '请输入验证码'
            }
        ]
    },
    {
        name: 'button',
        label: ''
    }
]
