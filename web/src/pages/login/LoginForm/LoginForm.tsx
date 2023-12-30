import { FC, useContext, useState, useEffect } from 'react'

import { Button, Form, Input } from 'antd'
import { useForm } from 'antd/es/form/Form'
import { useNavigate } from 'react-router-dom'
import { CopyrightOutlined } from '@ant-design/icons'
import DataForm from '@/components/Data/DataForm/DataForm'
import { formItems } from './options'
import { login } from '@/apis/login/login.api'
import { GlobalContext } from '@/context'
import { CaptchaRes, getCaptcha } from '@/apis/login/captcha.api'
import { AES_Encrypt } from '@/utils/aes'
import styles from '../style/login.module.less'

export type LoginParams = {
    username: string
    password: string
    code: string
}

const LoginForm: FC = () => {
    const { setUser, setToken } = useContext(GlobalContext)
    const [loginForm] = useForm<LoginParams>()
    const navigate = useNavigate()

    const [loading, setLoading] = useState(false)
    useEffect(() => {
        handleCaptcha()
    }, [])

    const handleLogin = () => {
        setLoading(true)
        loginForm.validateFields().then(({ username, password, code }) => {
            login({
                username,
                password: AES_Encrypt(password),
                code: code,
                captchaId: captcha.captchaId
            })
                .then((data) => {
                    setToken?.(data.token)
                    setUser?.(data.user)
                    navigate('/')
                })
                .catch(() => {
                    // 重新获取验证码
                    handleCaptcha()
                })
                .finally(() => {
                    setLoading(false)
                })
        })
    }

    const [captcha, setCaptcha] = useState<CaptchaRes>({
        captcha: '',
        captchaId: ''
    })

    const handleCaptcha = () => {
        getCaptcha({ x: 40, y: 160 })
            .then((res) => {
                setCaptcha(res)
            })
            .catch(() => {
                setCaptcha({
                    captcha: '',
                    captchaId: ''
                })
            })
    }

    return (
        <div className={styles.LoginForm}>
            <div>
                <div className={styles.LoginFormTitle}>登录</div>
                <DataForm
                    form={loginForm}
                    items={formItems}
                    formProps={{
                        layout: 'horizontal',
                        className: styles.LoginFormForm,
                        form: loginForm
                    }}
                    button={
                        <Button
                            type="primary"
                            size="large"
                            block
                            onClick={handleLogin}
                            loading={loading}
                        >
                            登录
                        </Button>
                    }
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
                                alt="验证码"
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
            </div>

            <div className={styles.LoginFormFooter}>
                <CopyrightOutlined />
                {window.location.host}
            </div>
        </div>
    )
}

export default LoginForm
