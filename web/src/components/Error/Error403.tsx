import { FC } from 'react'
import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'

const Error403: FC = () => {
    const navigate = useNavigate()

    const navigateToHome = () => {
        navigate('/home')
    }
    return (
        <Result
            status="403"
            title="403"
            subTitle="Sorry, you are not authorized to access this page."
            extra={
                <Button type="primary" onClick={navigateToHome}>
                    Back Home
                </Button>
            }
        />
    )
}

export default Error403
