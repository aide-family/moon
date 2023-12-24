import { useContext, FC } from 'react'
import { Breadcrumb } from 'antd'
import { useLocation } from 'react-router-dom'
import { GlobalContext } from '@/context'

const RouteBreadcrumb: FC = () => {
    const { breadcrumbNameMap } = useContext(GlobalContext)
    const location = useLocation()
    const pathSnippets = location.pathname.split('/').filter((i) => i)

    const extraBreadcrumbItems = pathSnippets.map((_, index) => {
        const url = `/${pathSnippets.slice(0, index + 1).join('/')}`
        return {
            key: url,
            title: breadcrumbNameMap?.[url]
        }
    })
    const breadcrumbItems = extraBreadcrumbItems
    return <Breadcrumb items={breadcrumbItems} />
}

export default RouteBreadcrumb
