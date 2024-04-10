import { FC } from 'react'
import { Result, Button } from 'antd'
import { useNavigate } from 'react-router-dom'

const Error404: FC = () => {
    const navigate = useNavigate()

    const navigateToHome = () => {
        navigate('/')
    }

    return (
        <Result
            status="404"
            title="404"
            subTitle="Sorry, the page you visited does not exist."
            extra={
                <Button type="primary" onClick={navigateToHome}>
                    Back Home
                </Button>
            }
        />
    )
}

export default Error404
