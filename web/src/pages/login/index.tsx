import React from 'react'
import LoginCarousel from './LoginCarousel/LoginCarousel'
import LoginForm from './LoginForm/LoginForm'
import styles from './style/login.module.less'
import { ThemeButton } from '@/components/ThemeButton'
import { Button, Space } from 'antd'
import { GithubButton } from '@/components/PromLayout/GithubButton'

const Login: React.FC = () => {
    return (
        <div className={styles.Login}>
            <Space size={8} className={styles.NavigationBar}>
                <GithubButton type="primary" />
                <Button type="primary" icon={<ThemeButton />} />
            </Space>

            <LoginCarousel />
            <LoginForm />
        </div>
    )
}

export default Login
