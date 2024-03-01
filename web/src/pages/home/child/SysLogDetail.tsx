import { SysLogActionTypeData } from '@/apis/data'
import { logApi } from '@/apis/home/system/log'
import { LogItem, defauleLogListReq } from '@/apis/home/system/log/types'
import { ModuleType, SysLogActionType } from '@/apis/types'
import { GlobalContext } from '@/context'
import {
    Button,
    Descriptions,
    Drawer,
    DrawerProps,
    Space,
    Timeline
} from 'antd'
import dayjs from 'dayjs'
import React, { useContext, useEffect, useState } from 'react'
import SyntaxHighlighter from 'react-syntax-highlighter'
import {
    atomOneDark,
    atomOneLight
} from 'react-syntax-highlighter/dist/esm/styles/hljs'

export interface SysLogDetailProps extends DrawerProps {
    module: ModuleType
    moduleId?: number
}

export const SysLogDetail: React.FC<SysLogDetailProps> = (props) => {
    const { onClose, module, moduleId } = props
    const { sysTheme } = useContext(GlobalContext)

    const [data, setData] = useState<LogItem[]>([])

    const fetchLogList = () => {
        if (!moduleId) return
        logApi
            .list({ ...defauleLogListReq, moduleName: module, moduleId })
            .then(({ list }) => {
                setData(list)
            })
    }

    useEffect(() => {
        fetchLogList()
    }, [moduleId, module])

    return (
        <Drawer
            {...props}
            extra={
                <Space size={8}>
                    {/* <Button type="primary">导出</Button> */}
                    <Button onClick={onClose}>关闭</Button>
                </Space>
            }
        >
            <Timeline
                mode="left"
                items={data.map((item) => {
                    const { text, color } = getAction(item.action)
                    return {
                        color: color,
                        // label: item.title,
                        children: (
                            <Descriptions
                                title={`[${text}] ${item.title}`}
                                column={2}
                                layout="vertical"
                                items={[
                                    {
                                        label: '操作人',
                                        key: 1,
                                        children: item?.user?.label || '-'
                                    },
                                    {
                                        label: '操作时间',
                                        key: 2,
                                        children: dayjs(
                                            +item.createdAt * 1000
                                        ).format('YYYY-MM-DD HH:mm:ss')
                                    },
                                    {
                                        label: '操作内容',
                                        key: 3,
                                        span: 2,
                                        children: (
                                            <div
                                                style={{
                                                    height: '400px',
                                                    overflowY: 'auto',
                                                    width: '100%'
                                                }}
                                            >
                                                <SyntaxHighlighter
                                                    language="json"
                                                    style={
                                                        sysTheme === 'dark'
                                                            ? atomOneDark
                                                            : atomOneLight
                                                    }
                                                >
                                                    {JSON.stringify(
                                                        JSON.parse(
                                                            item?.content ||
                                                                '{}'
                                                        ),
                                                        null,
                                                        2
                                                    )}
                                                </SyntaxHighlighter>
                                            </div>
                                        )
                                    }
                                ]}
                            />
                        )
                    }
                })}
            />
        </Drawer>
    )
}

const getAction = (action: SysLogActionType) => {
    return SysLogActionTypeData[action]
}
