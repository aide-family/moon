import userApi from '@/apis/home/system/user'
import { CaptchaRes, getCaptcha } from '@/apis/login/captcha.api'
import DataForm from '@/components/Data/DataForm/DataForm'
import { GlobalContext } from '@/context'
import { AES_Encrypt } from '@/utils/aes'
import { Form, Input, Modal, ModalProps } from 'antd'
import React, { useContext, useState } from 'react'

export interface ChangePassProps extends ModalProps {}

export type EditPassForm = {
    oldPass: string
    newPass: string
    confirmPass: string
    code: string
}

let timer: NodeJS.Timeout | null
export const ChangePass: React.FC<ChangePassProps> = (props) => {
    const { onOk } = props
    const { sysTheme } = useContext(GlobalContext)
    const [form] = Form.useForm<EditPassForm>()
    const [captcha, setCaptcha] = useState<CaptchaRes>({
        captcha: '',
        captchaId: ''
    })
    const [loading, setLoading] = useState(false)

    const oldPass: string = Form.useWatch('oldPass', form) || ''
    const newPass: string = Form.useWatch('newPass', form) || ''
    const confirmPass: string = Form.useWatch('confirmPass', form) || ''

    const handleCaptcha = () => {
        if (timer) {
            clearTimeout(timer)
        }
        timer = setTimeout(() => {
            form.validateFields().then(() => {
                getCaptcha({
                    x: 40,
                    y: 160,
                    captchaType: 3,
                    theme: sysTheme
                }).then(setCaptcha)
            })
        }, 200)
    }

    const handleOnOk = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
        setLoading(true)
        form.validateFields().then(({ oldPass, newPass, code }) => {
            userApi
                .userPasswordEdit({
                    oldPassword: AES_Encrypt(oldPass),
                    newPassword: AES_Encrypt(newPass),
                    code: code,
                    captchaId: captcha.captchaId
                })
                .then(() => onOk?.(e))
                .finally(() => {
                    setLoading(false)
                })
        })
    }

    return (
        <Modal {...props} onOk={handleOnOk} confirmLoading={loading}>
            <DataForm
                form={form}
                formProps={{ layout: 'vertical', form: form }}
                items={[
                    {
                        name: 'oldPass',
                        label: '旧密码',
                        formItemProps: undefined,
                        dataProps: {
                            type: 'password',
                            parentProps: {
                                placeholder: '请输入旧密码'
                            }
                        },
                        rules: [{ required: true, message: '请输入旧密码' }]
                    },
                    {
                        name: 'newPass',
                        label: '新密码',
                        formItemProps: undefined,
                        dataProps: {
                            type: 'password',
                            parentProps: {
                                placeholder: '请输入新密码'
                            }
                        },
                        rules: [
                            { required: true, message: '请输入新密码' },
                            { min: 6, message: '密码长度不能小于6位' },
                            {
                                validator: (_, value: string) => {
                                    if (!value || value.length < 6) {
                                        return Promise.resolve()
                                    }
                                    if (
                                        value &&
                                        value.length >= oldPass.length &&
                                        value === oldPass
                                    ) {
                                        return Promise.reject(
                                            '新密码不能与旧密码相同'
                                        )
                                    }
                                    // 如果confirmPass有值，也要判断必须相同
                                    if (
                                        confirmPass &&
                                        confirmPass !== newPass
                                    ) {
                                        return Promise.reject('两次密码不一致')
                                    }
                                    return Promise.resolve()
                                }
                            }
                        ]
                    },
                    {
                        name: 'confirmPass',
                        label: '确认密码',
                        formItemProps: undefined,
                        dataProps: {
                            type: 'password',
                            parentProps: {
                                placeholder: '请输入确认密码'
                            }
                        },
                        rules: [
                            { required: true, message: '请输入确认密码' },
                            { min: 6, message: '密码长度不能小于6位' },
                            {
                                validator: (_, value: string) => {
                                    if (!value || value.length < 6) {
                                        return Promise.resolve()
                                    }
                                    if (
                                        value &&
                                        newPass &&
                                        value.length >= newPass.length &&
                                        value !== newPass
                                    ) {
                                        return Promise.reject('两次密码不一致')
                                    }
                                    return Promise.resolve()
                                }
                            }
                        ]
                    },
                    {
                        name: 'code',
                        label: '',
                        formItemProps: {
                            required: true
                        }
                    }
                ]}
                code={
                    <div
                        style={{
                            display: 'flex',
                            gap: 8,
                            height: 40
                        }}
                    >
                        <Input size="large" placeholder="验证码" />
                        <img
                            src={captcha.captcha}
                            // alt="点击获取验证码"
                            width={160}
                            height={40}
                            style={{
                                boxSizing: 'border-box',
                                border: '1px solid #ccc',
                                cursor: 'pointer',
                                borderRadius: 8
                            }}
                            onClick={handleCaptcha}
                        />
                    </div>
                }
            />
        </Modal>
    )
}
