import React, {useEffect, useRef, useState} from 'react'
import RouteBreadcrumb from "@/components/PromLayout/RouteBreadcrumb";
import {DataOption, DataTable, SearchForm} from "@/components/Data";
import type {ColumnGroupType, ColumnType} from "antd/es/table";
import {ListEndpointRequest, PrometheusServerItem} from "@/apis/home/monitor/endpoint/types.ts";
import dayjs from "dayjs";
import {ActionKey} from "@/apis/data.ts";
import {Form} from "antd";
import {options} from "./options.tsx"
import {HeightLine, PaddingLine} from "@/components/HeightLine";

type EndpointColumnType = ColumnType<PrometheusServerItem> | ColumnGroupType<PrometheusServerItem>

let timer: NodeJS.Timeout
const {leftOptions, rightOptions, searchItems} = options
const defaultPadding = 12

const Endpoint: React.FC = () => {
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<PrometheusServerItem[]>([])
    const [refresh, setRefresh] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(false)
    const [search, setSearch] = useState<ListEndpointRequest>({
        page: {
            curr: 1,
            size: 10
        },
        keyword: ''
    })
    const columns: EndpointColumnType[] = [
        {
            title: '数据源名称',
            dataIndex: 'name',
            key: 'name',
            width: 220
        },
        {
            title: '端点',
            dataIndex: 'endpoint',
            key: 'endpoint',
            width: 220
        },
        {
            title: '端点状态',
            dataIndex: 'status',
            key: 'status',
            width: 220
        },
        {
            title: '备注',
            dataIndex: 'remark',
            key: 'remark',
            ellipsis: true
        },
        {
            title: '创建时间',
            dataIndex: 'createdAt',
            key: 'createdAt',
            width: 220,
            render: (createdAt: number | string) => {
                return dayjs(createdAt).format('YYYY-MM-DD HH:mm:ss')
            }
        },
        {
            title: '更新时间',
            dataIndex: 'updatedAt',
            key: 'updatedAt',
            width: 220,
            render: (updatedAt: number | string) => {
                return dayjs(updatedAt).format('YYYY-MM-DD HH:mm:ss')
            }
        }
    ]

    const handlerGetData = () => {
        setLoading(true)
        // const {data} = await endpointApi.list()
        setDataSource([])
        setLoading(false)
    }

    // 刷新
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
    }

    //操作栏按钮
    const handleOptionClick = (val: ActionKey) => {
        switch (val) {
            case ActionKey.ADD:
                break
            case ActionKey.REFRESH:
                handlerRefresh()
                break
            case ActionKey.RESET:
                setSearch({
                    keyword: '',
                    page: {
                        curr: 1,
                        size: 10
                    }
                })
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (
        changedValues: any, // TODO 不要any
        allValues: any
    ) => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            setSearch({
                ...search,
                ...changedValues
            })
            console.log(changedValues, allValues)
        }, 500)
    }


    useEffect(() => {
        handlerGetData()
    }, [refresh, search])

    return (
        <div className="bodyContent">
            <div ref={operationRef}>
                <RouteBreadcrumb/>
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
                    rightOptions={rightOptions(loading)}
                    leftOptions={leftOptions}
                    action={handleOptionClick}
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
                operationRef={operationRef}
                // total={total}
                loading={loading}
                // operationItems={operationItems}
                // pageOnChange={handlerTablePageChange}
                // rowSelection={{
                //     onChange: handlerBatchData,
                //     selectedRowKeys: tableSelectedRows.map((item) => item.id)
                // }}
                // action={handlerTableAction}
            />
        </div>
    )
}

export default Endpoint