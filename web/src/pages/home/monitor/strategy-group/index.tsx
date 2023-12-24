import React, { Key, useEffect, useRef, useState } from 'react'
import { Button, Form } from 'antd'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption.tsx'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { Detail } from '@/pages/home/monitor/strategy-group/child/detail.tsx'
import { StrategyGroupItemType } from '@/pages/home/monitor/strategy-group/type.ts'
import { defaultData } from '@/pages/home/monitor/strategy-group/data.ts'
import {
    OP_KEY_STRATEGY_LIST,
    searchItems,
    tableOperationItems
} from '@/pages/home/monitor/strategy-group/options.tsx'
import { DictSelectItem } from '@/apis/home/system/dict/types.ts'
import { ActionKey } from '@/apis/data.ts'

const defaultPadding = 12

const StrategyGroup: React.FC = () => {
    const navigate = useNavigate()
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<StrategyGroupItemType[]>([])

    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)

    const columns: (
        | ColumnGroupType<StrategyGroupItemType>
        | ColumnType<StrategyGroupItemType>
    )[] = [
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
            width: 160,
            render: (name: string) => {
                return name
            }
        },
        {
            title: '描述',
            dataIndex: 'remark',
            key: 'remark',
            width: 160,
            render: (description: string) => {
                return description
            }
        },
        {
            title: '类型',
            dataIndex: 'categories',
            key: 'categories',
            width: 160,
            render: (categories?: DictSelectItem[]) => {
                return categories?.map((item: DictSelectItem) => {
                    return item?.label
                })
            }
        },
        {
            title: '策略组状态',
            dataIndex: 'status',
            key: 'status',
            width: 160,
            render: (status: string) => {
                return status
            }
        },
        {
            title: '策略组创建时间',
            dataIndex: 'created_at',
            key: 'created_at',
            width: 160,
            render: (created_at: string) => {
                return created_at
            }
        },
        {
            title: '策略组更新时间',
            dataIndex: 'updated_at',
            key: 'updated_at',
            width: 160,
            render: (updated_at: string) => {
                return updated_at
            }
        }
    ]

    const handlerOpenDetail = () => {
        setOpenDetail(true)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
    }

    // 获取数据
    const handlerGetData = () => {
        setLoading(true)
        setTimeout(() => {
            setDataSource(defaultData)
            setTotal(203)
            setLoading(false)
        }, 500)
    }

    useEffect(() => {
        handlerGetData()
    }, [refresh])

    // 刷新
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
    }

    // 分页变化
    const handlerTablePageChange = (page: number, pageSize?: number) => {
        console.log(page, pageSize)
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: Key[],
        selectedRows: StrategyGroupItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    const toStrategyListPage = (record: StrategyGroupItemType) => {
        console.log(record)
        navigate(`/home/monitor/strategy-group/strategy?groupId=${record.id}`)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: string, record: StrategyGroupItemType) => {
        console.log(key, record)
        switch (key) {
            case OP_KEY_STRATEGY_LIST:
                toStrategyListPage(record)
                break
            case ActionKey.DETAIL:
                handlerOpenDetail()
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (
        changedValues: any,
        allValues: any
    ) => {
        console.log(changedValues, allValues)
    }

    const leftOptions: DataOptionItem[] = [
        {
            key: '批量导入',
            label: (
                <Button type="primary" loading={loading}>
                    批量导入
                </Button>
            )
        }
    ]

    const rightOptions: DataOptionItem[] = [
        {
            key: 'refresh',
            label: (
                <Button
                    type="primary"
                    loading={loading}
                    onClick={handlerRefresh}
                >
                    刷新
                </Button>
            )
        }
    ]

    return (
        <div className="bodyContent">
            <Detail open={openDetail} onClose={handlerCloseDetail} id="1" />
            <div ref={operationRef}>
                <RouteBreadcrumb />
                <HeightLine />
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearFormValuesChange
                    }}
                />
                <HeightLine />
                <DataOption
                    queryForm={queryForm}
                    rightOptions={rightOptions}
                    leftOptions={leftOptions}
                />
                <PaddingLine
                    padding={defaultPadding}
                    height={1}
                    borderRadius={4}
                />
            </div>

            <DataTable
                dataSource={dataSource}
                columns={columns}
                total={total}
                loading={loading}
                operationItems={tableOperationItems}
                pageOnChange={handlerTablePageChange}
                rowSelection={{
                    onChange: handlerBatchData
                }}
                action={handlerTableAction}
            />
        </div>
    )
}

export default StrategyGroup
