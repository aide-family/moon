import { useContext, useEffect, useState, RefObject, FC, useRef } from 'react'

import type { MenuProps, TableProps } from 'antd'
import type { ColumnGroupType, ColumnType } from 'antd/es/table'
import { Table, ConfigProvider, Button, Space, Tooltip } from 'antd'

import { GlobalContext } from '@/context'
import { IconFont } from '@/components/IconFont/IconFont'
import { MoreMenu } from '../'
import { ActionKey } from '@/apis/data'
import { RoleListItem } from '@/apis/home/system/role/types.ts'
import type { Reference } from 'rc-table'
import { random } from '@/utils/random'

export type DataTableProps<T = any> = TableProps<T> & {
    // 是否显示序号
    showIndex?: boolean
    // 是否显示操作列
    showOperation?: boolean
    // 操作区Ref
    operationRef?: RefObject<HTMLDivElement>
    defaultPadding?: number
    operationItems?: (item?: T) => MenuProps['items']
    action?: (key: ActionKey, record: any) => void
    // 数据分页
    total?: number
    pageSize?: number
    current?: number
    pageOnChange?: (page: number, pageSize: number) => void
    x?: number | string
    y?: number | string
}

const defaultIndexColumn = (
    current: number,
    pageSize: number
): ColumnGroupType<any> | ColumnType<any> => ({
    title: '序号',
    key: 'index',
    width: 80,
    fixed: 'left',
    render: (_text: any, _record: any, index: number) => {
        return <span>{(current - 1) * pageSize + index + 1}</span>
    }
})

const defaultOperation = (
    items: (item?: RoleListItem) => MenuProps['items'],
    action?: (key: ActionKey, record: any) => void
): ColumnGroupType<any> | ColumnType<any> => ({
    title: '操作',
    key: ActionKey.ACTION,
    width: 120,
    fixed: 'right',
    align: 'center',
    render: (_, record: any) => {
        return (
            <Space>
                <Tooltip title="详情">
                    <Button
                        size="large"
                        type="link"
                        icon={<IconFont type="icon-xiangqing" />}
                        onClick={() => action?.(ActionKey.DETAIL, record)}
                    />
                </Tooltip>
                {items && items?.length > 0 && (
                    <MoreMenu
                        items={items(record)}
                        onClick={(key) =>
                            action?.(key as ActionKey.ACTION, record)
                        }
                    />
                )}
            </Space>
        )
    }
})
let layoutContentElement = document.getElementById('root')
let timer: NodeJS.Timeout | null = null
const DataTable: FC<DataTableProps> = (props) => {
    const { size } = useContext(GlobalContext)
    const {
        dataSource = [],
        columns = [],
        showIndex = true,
        showOperation = true,
        operationItems = () => [],
        action,
        total,
        pageSize,
        current,
        pageOnChange,
        x = '100vw',
        y
    } = props

    const tableRef = useRef<Reference>(null)
    const [tableScrollHeight, setTableScrollHeight] = useState<number>()

    const [_columns, setColumns] = useState<
        (ColumnGroupType<any> | ColumnType<any>)[]
    >([])

    useEffect(() => {
        let columnsTmp = columns
        if (showIndex) {
            columnsTmp = [
                defaultIndexColumn(current || 1, pageSize || 10),
                ...columns
            ]
        }
        if (showOperation) {
            // 判断最后一个是否为操作列
            if (columnsTmp[columnsTmp.length - 1]?.key === ActionKey.ACTION) {
                columnsTmp[columnsTmp.length - 1] = {
                    ...columnsTmp[columnsTmp.length - 1],
                    fixed: 'right'
                }
            } else {
                columnsTmp.push(defaultOperation(operationItems, action))
            }
        }
        // 根据key去重
        const uniqueItems = Array.from(
            new Map(columnsTmp.map((item) => [item.key, item])).values()
        )
        setColumns([...uniqueItems])
    }, [columns, pageSize, current])

    const getScrollHeight = (height: number) => {
        let header =
            document.getElementsByClassName('ant-layout-header')[0]
                ?.clientHeight
        let footer =
            document.getElementsByClassName('ant-layout-footer')?.[0]
                ?.clientHeight
        let tableHeader =
            document.getElementsByClassName('ant-table-thead')[0]?.clientHeight
        let tablePage =
            document.getElementsByClassName('ant-table-pagination')[0]
                ?.clientHeight || 32
        let searchBox = tableRef.current?.nativeElement?.offsetTop || 0

        let scrollHeight =
            height -
            (header +
                footer +
                tableHeader +
                tablePage +
                searchBox +
                8 * 2 +
                16 * 2)
        if (timer) {
            clearTimeout(timer)
        }
        timer = setTimeout(() => {
            setTableScrollHeight(scrollHeight)
        }, 500)
    }

    const resizeObserver = new ResizeObserver((entries) => {
        for (const entry of entries) {
            const { height } = entry.contentRect
            getScrollHeight(height)
        }
    })

    useEffect(() => {
        if (tableRef) {
            resizeObserver.observe(layoutContentElement!)
        }
        return () => {
            resizeObserver.disconnect()
        }
    }, [layoutContentElement, size])

    return (
        <>
            <ConfigProvider>
                <Table
                    {...props}
                    ref={tableRef}
                    rowKey={(record) => `${record?.id || random(-100000, -1)}`}
                    dataSource={dataSource}
                    columns={_columns.map((item, index) => {
                        if (index === 0) {
                            item.fixed = true
                        }
                        return item
                    })}
                    scroll={{ x: x, y: y || tableScrollHeight }}
                    size={size}
                    sticky
                    pagination={{
                        total: total,
                        showTotal: (total) => `共 ${total} 条`,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        onChange: pageOnChange,
                        pageSize: pageSize,
                        current: current
                    }}
                />
            </ConfigProvider>
        </>
    )
}

export default DataTable
