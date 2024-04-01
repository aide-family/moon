import { useContext, FC, useState, useEffect } from 'react'
import { Breadcrumb } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { GlobalContext } from '@/context'
import useToken from 'antd/es/theme/useToken'

let timer: NodeJS.Timeout | null = null
const RouteBreadcrumb: FC = () => {
    const { breadcrumbNameMap } = useContext(GlobalContext)
    const location = useLocation()
    const pathSnippets = location.pathname.split('/').filter((i) => i)
    const navigate = useNavigate()
    const [, token] = useToken()
    const [title, setTitle] = useState('Moon')

    const extraBreadcrumbItems = pathSnippets.map((_, index) => {
        const url = `/${pathSnippets.slice(0, index + 1).join('/')}`
        const breadcrumbName = breadcrumbNameMap?.[url]
        const disabled =
            breadcrumbName?.disabled || index === pathSnippets.length - 1
        if (index === pathSnippets.length - 1) {
            if (timer) {
                clearTimeout(timer)
            }
            timer = setTimeout(() => {
                setTitle(breadcrumbName?.name || 'Moon')
            }, 100)
        }
        return {
            key: url,
            title: (
                <a
                    style={{
                        color: disabled ? '' : token['blue-6'],
                        cursor: disabled ? 'no-drop' : 'pointer'
                    }}
                    onClick={() => {
                        navigate(url)
                    }}
                >
                    {breadcrumbName?.name}
                </a>
            )
        }
    })
    const breadcrumbItems = extraBreadcrumbItems

    useEffect(() => {
        if (!title) return
        document.title = title
    }, [title])
    return <Breadcrumb items={breadcrumbItems} />
}

export default RouteBreadcrumb
