import {
    Suspense,
    useContext,
    useEffect,
    useState,
    FC,
    MouseEvent
} from 'react'

import { Button, Layout, Space, Watermark, theme } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { CopyrightOutlined } from '@ant-design/icons'

import { GlobalContext } from '@/context'
import Loading from '../Loading'
import HeaderTitle from './HeaderTitle/HeaderTitle'
import { SiderMenu } from './SiderMenu/SiderMenu'
import { UserInfo } from './UserInfo/UserInfo'
import { Msg } from './Msg/Msg'
import { Setting } from './Setting/Setting'

import styles from './style/index.module.less'
import { refreshToken } from '@/apis/login/login.api'
import { ThemeButton } from '../ThemeButton'
import { GithubButton } from './GithubButton'

const { Header, Footer, Sider, Content } = Layout
const { useToken } = theme

export type PromLayoutProps = {
    watermark?: string
}

export const LayoutContentID = 'LayoutContent'

const PromLayout: FC<PromLayoutProps> = (props) => {
    const { token } = useToken()
    const {
        user,
        setLayoutContentElement,
        setAuthToken,
        autToken,
        setIntervalId
    } = useContext(GlobalContext)
    const navigator = useNavigate()
    const { watermark = user?.username } = props
    const [collapsed, setCollapsed] = useState(true)

    const handleOnMouseEnter = (e: MouseEvent) => {
        // 鼠标移入, 展开菜单
        if (e.type === 'mouseenter') {
            setCollapsed(false)
        }
        // 鼠标移出, 收起菜单
        if (e.type === 'mouseleave') {
            setCollapsed(true)
        }
    }

    useEffect(() => {
        setLayoutContentElement?.(document.getElementById(LayoutContentID))
    }, [])

    const local = useLocation()

    useEffect(() => {
        // TODO 做路由权限认证
        console.log('TODO 做路由权限认证', local)
        if (!autToken) {
            navigator('/login')
        }
    }, [local.pathname])

    const handleRefreshToken = () => {
        refreshToken().then((data) => {
            setAuthToken?.(data.token)
        })
    }

    useEffect(() => {
        const id = setInterval(() => {
            handleRefreshToken()
        }, 1000 * 60 * 10) // 10分钟
        setIntervalId?.(id)
    }, [])

    return (
        <Layout className={[styles.widthHight100, styles.Layout].join(' ')}>
            <Header className={styles.LayoutHeader}>
                <HeaderTitle />
                <Space size={12} direction="horizontal">
                    {/*<SpaceInfo />*/}
                    <GithubButton />
                    <Msg />
                    <Button
                        style={{
                            color: '#FFF'
                        }}
                        type="text"
                        icon={<ThemeButton />}
                    />
                    <Setting />
                    <UserInfo />
                </Space>
            </Header>
            <Layout>
                <Sider
                    defaultCollapsed
                    collapsed={collapsed}
                    className={styles.LayoutSider}
                    onMouseEnter={handleOnMouseEnter}
                    onMouseLeave={handleOnMouseEnter}
                    collapsedWidth={60}
                >
                    <Watermark
                        content={watermark}
                        font={{ color: token.colorBgTextHover }}
                        className="wh100"
                    >
                        <SiderMenu inlineCollapsed={collapsed} />
                    </Watermark>
                </Sider>

                <Layout>
                    <Content>
                        <Suspense fallback={<Loading />}>
                            <div
                                className={styles.LayoutContent}
                                id={LayoutContentID}
                            >
                                <Watermark
                                    content={watermark}
                                    className="wh100"
                                    font={{ color: token.colorBgTextHover }}
                                >
                                    <div
                                        className="bodyContent"
                                        style={{
                                            background: token.colorBgContainer,
                                            color: token.colorTextBase
                                        }}
                                    >
                                        <Outlet />
                                    </div>
                                </Watermark>
                            </div>
                        </Suspense>
                    </Content>
                    <Footer
                        className={styles.LayoutFooter}
                        style={{
                            background: token.colorBgContainer,
                            color: token.colorTextBase
                        }}
                    >
                        <CopyrightOutlined />
                        {window.location.host}
                    </Footer>
                </Layout>
            </Layout>
        </Layout>
    )
}

export default PromLayout
