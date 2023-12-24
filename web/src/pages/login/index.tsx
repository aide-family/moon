import React from 'react'
import LoginCarousel from './LoginCarousel/LoginCarousel'
import LoginForm from './LoginForm/LoginForm'
import styles from './style/login.module.less'

const Login: React.FC = () => {
    return (
        <div className={styles.Login}>
            <LoginCarousel />
            <LoginForm />
        </div>
    )
}

export default Login
