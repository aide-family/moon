import { FC } from 'react'
import { Button, Badge } from 'antd'
import { SoundOutlined } from '@ant-design/icons'

const Msg: FC = () => {
    return (
        <>
            <Badge count={1} size="small" offset={[-5, 8]}>
                <Button
                    type="text"
                    style={{
                        color: '#FFF'
                    }}
                    icon={<SoundOutlined />}
                />
            </Badge>
        </>
    )
}

export { Msg }
