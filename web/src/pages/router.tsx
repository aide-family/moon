import { lazy } from 'react'
import type { RouteObject } from 'react-router-dom'
import { Navigate } from 'react-router-dom'
import { Error404, Error403 } from '@/components/Error'

export const routers: RouteObject[] = [
    {
        path: '/home',
        Component: lazy(() => import('@/components/PromLayout')),
        children: [
            {
                path: '/home',
                Component: lazy(() => import('@/pages/home'))
            },
            {
                path: '/home/monitor/strategy-group',
                Component: lazy(
                    () => import('@/pages/home/monitor/strategy-group')
                ),
                children: []
            },
            {
                path: '/home/monitor/strategy-group/strategy',
                Component: lazy(() => import('@/pages/home/monitor/strategy'))
            },
            {
                path: '/home/monitor/endpoint',
                Component: lazy(() => import('@/pages/home/monitor/endpoint'))
            },
            {
                path: '/home/system/user',
                Component: lazy(() => import('@/pages/home/system/user'))
            },
            {
                path: '/home/system/dict',
                Component: lazy(() => import('@/pages/home/system/dict'))
            },
            {
                path: '/home/system/role',
                Component: lazy(() => import('@/pages/home/system/role'))
            },
            {
                path: '/home/system/menu',
                Component: lazy(() => import('@/pages/home/system/menu'))
            },
            {
                path: '/home/system/auth',
                Component: lazy(() => import('@/pages/home/system/auth'))
            },
            {
                path: '/home/customer/list',
                Component: lazy(() => import('@/pages/home/customer/customer'))
            },
            {
                path: '/home/customer/supplier',
                Component: lazy(() => import('@/pages/home/customer/supplier'))
            },
            {
                path: '/home/resource/device',
                Component: lazy(() => import('@/pages/home/resource/device'))
            },
            {
                path: '/home/resource/node',
                Component: lazy(() => import('@/pages/home/resource/node'))
            },
            {
                path: '/home/resource/account',
                Component: lazy(() => import('@/pages/home/resource/account'))
            },
            {
                // 404
                path: '*',
                element: <Error404 />
            }
        ]
    },
    {
        path: '/login',
        Component: lazy(() => import('@/pages/login'))
    },
    {
        path: '/',
        // 重定向/home
        element: <Navigate to="/home" replace={true} />
    },
    {
        // 403
        path: '*',
        element: <Error403 />
    }
]
