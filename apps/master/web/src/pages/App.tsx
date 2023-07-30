import React, {lazy, Suspense, useEffect, useState} from 'react'
import {HashRouter, Route, Routes} from 'react-router-dom'

import {GlobalContext, GlobalContextType, SpanceInfoType} from './context'
import changeTheme from '../utils/changeTheme'
import {IRoute, routes} from "@/config/routes";
import './App.css'

type ThemeType = 'dark' | 'light'
type GlobalData = {
    spaceId?: string
    theme?: ThemeType
}

declare global {
    interface Window {
        __MICRO_APP_BASE_ROUTE__: any
        microApp: any
    }
}

function App() {
    const [spaceId, setSpaceId] = useState('')
    const [theme, setTheme] = useState<ThemeType>(
        (localStorage.getItem('theme') as ThemeType) || 'light'
    )

    const contextValue: GlobalContextType = {
        spaceId,
        setSpaceId,
        spaceInfo: JSON.parse(
            localStorage.getItem('spaceInfo') || `{"id": "","name": ""}`
        ) as SpanceInfoType,
        theme,
    }

    const renderRoutes = (routes: IRoute[]) => {
        const ignoreRoutes: React.ReactElement[] = []

        let index = 0
        const render = (route: IRoute, parentPath = '') => {
            index++

            if (route.children && route.children.length > 0) {
                route.children.forEach((item) => {
                    render(item, route.path)
                })
            }

            let routePath = route.path || ''
            if (parentPath) {
                routePath = `/${parentPath}/${route?.path}`.replace(/\/+/g, '/')
            }
            let routeDom: React.ReactElement = (
                <Route
                    key={`${index}`}
                    path={routePath}
                    Component={lazy(() => import(`@/pages${routePath}`))}
                />
            )

            ignoreRoutes.push(routeDom)
        }

        routes.forEach((route) => {
            return render(route)
        })

        return <Route>{ignoreRoutes}</Route>
    }


    useEffect(() => {
        window.microApp?.addGlobalDataListener(({spaceId, theme}: GlobalData) => {
            if (spaceId) {
                setSpaceId(spaceId)
            }
            if (theme) {
                setTheme(theme)
            }
        }, false)
    }, [])

    useEffect(() => {
        changeTheme(theme)
    }, [theme])

    return (
        <HashRouter basename={window.__MICRO_APP_BASE_ROUTE__ || '/'}>
            <GlobalContext.Provider value={contextValue}>
                <Suspense fallback={<div>Loading...</div>}>
                    <Routes>{renderRoutes(routes)}</Routes>
                </Suspense>
            </GlobalContext.Provider>
        </HashRouter>
    )
}

export default App
