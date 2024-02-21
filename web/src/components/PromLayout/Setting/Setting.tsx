import { FC } from 'react'

import { Button } from 'antd'
import { IconFont } from '@/components/IconFont/IconFont'

const Setting: FC = () => {
    return (
        <>
            <Button
                style={{
                    color: '#FFF'
                }}
                type="text"
                icon={<IconFont type="icon-configure" />}
            />
        </>
    )
}

export { Setting }
