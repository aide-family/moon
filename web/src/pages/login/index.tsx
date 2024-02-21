import React from 'react'
import LoginCarousel from './LoginCarousel/LoginCarousel'
import LoginForm from './LoginForm/LoginForm'
import styles from './style/login.module.less'
import { ThemeButton } from '@/components/ThemeButton'
import { Button } from 'antd'

const Login: React.FC = () => {
    return (
        <div className={styles.Login}>
            <Button
                type="primary"
                icon={<ThemeButton />}
                className={styles.ThemeButton}
            />
            <LoginCarousel />
            <LoginForm />
        </div>
    )
}

export default Login
