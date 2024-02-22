import { GithubOutlined } from '@ant-design/icons'
import { Button, ButtonProps } from 'antd'
import React from 'react'

export interface GihubButtonProps extends ButtonProps {}

export const GithubButton: React.FC<GihubButtonProps> = (props) => {
    const linkGithub = () => {
        // 打开新页面跳转
        window.open('https://github.com/aide-family/moon')
    }

    return (
        <Button
            type="text"
            onClick={linkGithub}
            icon={<GithubOutlined />}
            style={{
                color: '#FFF'
            }}
            {...props}
        />
    )
}
