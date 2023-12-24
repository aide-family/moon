import { FC } from 'react'
import { Button, Badge } from 'antd'
import { IconFont } from '@/components/IconFont/IconFont'

const Msg: FC = () => {
    return (
        <>
            <Badge count={1} size="small" offset={[-5, 8]}>
                <Button
                    type="text"
                    icon={<IconFont type="icon-message-fill" />}
                />
            </Badge>
        </>
    )
}

export { Msg }
