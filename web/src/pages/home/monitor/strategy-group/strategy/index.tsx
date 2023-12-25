import {FC, Key, useEffect, useRef, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import {Button, Form} from 'antd'
import {ColumnGroupType, ColumnType} from 'antd/es/table'
import {DataOptionItem} from '@/components/Data/DataOption/DataOption.tsx'
import {Detail} from '@/pages/home/monitor/strategy-group/child/detail.tsx'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import {HeightLine, PaddingLine} from '@/components/HeightLine'
import {DataOption, DataTable, SearchForm} from '@/components/Data'
import {StrategyItemType} from '@/pages/home/monitor/strategy-group/strategy/type.ts'
import {
    OP_KEY_STRATEGY_GROUP_LIST,
    searchItems,
    tableOperationItems
} from '@/pages/home/monitor/strategy-group/strategy/options.tsx'
import {defaultData} from '@/pages/home/monitor/strategy-group/strategy/data.ts'
import {CopyOutlined} from '@ant-design/icons'
import {ActionKey} from "@/apis/data.ts";

const defaultPadding = 12

const Strategy: FC = () => {
    const navigate = useNavigate()
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<StrategyItemType[]>([])

    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)

    const columns: (
        | ColumnGroupType<StrategyItemType>
        | ColumnType<StrategyItemType>
        )[] = [
        {
            title: '名称',
            dataIndex: 'alert',
            key: 'alert',
            width: 160,
            render: (alert: string) => {
                return alert
            }
        },
        {
            title: '持续时间',
            dataIndex: 'duration',
            key: 'duration',
            width: 160,
            render: (duration: string) => {
                return duration
            }
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            width: 160,
            render: (status: string) => {
                return status
            }
        },
        {
            title: '创建时间',
            dataIndex: 'created_at',
            key: 'created_at',
            width: 160,
            render: (created_at: string) => {
                return created_at
            }
        },
        {
            title: '更新时间',
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
        selectedRows: StrategyItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    const toStrategyGroupPage = (record: StrategyItemType) => {
        console.log(record)
        navigate(`/home/monitor/strategy-group?id=${record.id}`)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: string, record: StrategyItemType) => {
        console.log(key, record)
        switch (key) {
            case OP_KEY_STRATEGY_GROUP_LIST:
                toStrategyGroupPage(record)
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
            <Detail open={openDetail} onClose={handlerCloseDetail} id="1"/>
            <div ref={operationRef}>
                <RouteBreadcrumb/>
                <HeightLine/>
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearFormValuesChange
                    }}
                />
                <HeightLine/>
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
                // operationRef={operationRef}
                total={total}
                loading={loading}
                operationItems={tableOperationItems}
                pageOnChange={handlerTablePageChange}
                rowSelection={{
                    onChange: handlerBatchData
                }}
                action={handlerTableAction}
                expandable={{
                    expandedRowRender: (record: StrategyItemType) => (
                        <div>
                            <CopyOutlined/>
                            <p style={{margin: 0}}>{record.expr}</p>
                        </div>
                    )
                }}
            />
        </div>
    )
}

export default Strategy
