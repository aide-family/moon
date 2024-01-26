import { useContext, useEffect, useState, RefObject, FC } from 'react'

import type { MenuProps, TableProps } from 'antd'
import type { ColumnGroupType, ColumnType } from 'antd/es/table'
import { Table, ConfigProvider, Button, Space, Tooltip } from 'antd'

import { GlobalContext } from '@/context'
import { IconFont } from '@/components/IconFont/IconFont'
import { MoreMenu } from '../'
import { ActionKey } from '@/apis/data'
import { RoleListItem } from '@/apis/home/system/role/types.ts'

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

const DataTable: FC<DataTableProps> = (props) => {
    const { size } = useContext(GlobalContext)
    const {
        dataSource = [],
        columns = [],
        showIndex = true,
        showOperation = true,
        // operationRef,
        // defaultPadding = 12,
        operationItems = () => [],
        action,
        total,
        pageSize,
        current,
        pageOnChange,
        x = 1500,
        y = 480
    } = props
    const [_columns, setColumns] = useState<
        (ColumnGroupType<any> | ColumnType<any>)[]
    >([])
    // const [tableHigh, setTableHigh] = useState<number>(
    //     (layoutContentElement?.clientHeight || 20) -
    //         (operationRef?.current?.clientHeight || 20)
    // )

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
        setColumns([...columnsTmp])
    }, [columns, pageSize, current])

    // const resizeObserver = new ResizeObserver((entries) => {
    //     for (const entry of entries) {
    //         const { height } = entry.contentRect
    //         let paginationHeight =
    //             document.getElementsByClassName('ant-table-pagination')?.[0]
    //                 ?.clientHeight || 30
    //         setTableHigh(
    //             height -
    //                 (operationRef?.current?.clientHeight || 20) -
    //                 paginationHeight
    //         )
    //     }
    // })

    // useEffect(() => {
    //     if (layoutContentElement) {
    //         resizeObserver.observe(layoutContentElement)
    //     }
    //     return () => {
    //         resizeObserver.disconnect()
    //     }
    // }, [layoutContentElement, size, operationRef])

    return (
        <>
            <ConfigProvider>
                <Table
                    {...props}
                    rowKey={(record) => record?.id}
                    dataSource={dataSource}
                    columns={_columns}
                    // scroll={{ x: 1500, y: tableHigh - defaultPadding * 8 }}
                    scroll={{ x: x, y: y }}
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
