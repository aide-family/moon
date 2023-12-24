import React, { useEffect, useState } from 'react'

import { Button, Form } from 'antd'
import type { ColumnType, ColumnGroupType } from 'antd/es/table'

import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { operationItems } from '@/components/Data/DataOption/option'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'

import { searchItems } from './options'
import { defaultData } from './data'
import Detail from './child/Detail'
import type { SupplierItemType } from './type'
import { ActionKey } from '@/apis/data'

const defaultPadding = 12

const Supplier: React.FC = () => {
    const oprationRef = React.useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const columns: (
        | ColumnGroupType<SupplierItemType>
        | ColumnType<SupplierItemType>
    )[] = [
        {
            title: '公司名称',
            dataIndex: 'company',
            key: 'company',
            fixed: 'left',
            width: 220
        },
        {
            title: '供应商类型',
            dataIndex: 'type',
            key: 'type',
            width: 120
        },
        {
            title: '联系人',
            dataIndex: 'contacts',
            key: 'contacts',
            width: 120
        },
        {
            title: '联系电话',
            dataIndex: 'phone',
            key: 'phone',
            width: 120
        },
        {
            title: '微信交流群',
            dataIndex: 'chat_group',
            key: 'chat_group',
            width: 220
        },
        {
            title: '联系地址',
            dataIndex: 'address',
            key: 'address',
            width: 220
        }
    ]
    const [dataSource, setDataSource] = useState<SupplierItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [supplierDetail, setSupplierDetail] = useState<
        SupplierItemType | undefined
    >()

    const handlerOpenDetail = () => {
        setOpenDetail(true)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
        setSupplierDetail(undefined)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: SupplierItemType) => {
        console.log(key, record)
        switch (key) {
            case ActionKey.DETAIL:
                handlerOpenDetail()
                setSupplierDetail(record)
                break
        }
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: React.Key[],
        selectedRows: SupplierItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
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

    useEffect(() => {
        setLoading(true)
        setTimeout(() => {
            setDataSource(defaultData)
            setTotal(defaultData.length)
            setLoading(false)
        }, 500)
    }, [refresh])

    return (
        <div className="bodyContent">
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                supplierId={supplierDetail?.id}
            />
            <div ref={oprationRef}>
                <RouteBreadcrumb />
                <HeightLine />
                <SearchForm form={queryForm} items={searchItems} />
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
                columns={columns}
                loading={loading}
                oprationRef={oprationRef}
                dataSource={dataSource}
                operationItems={operationItems}
                total={total}
                rowSelection={{
                    onChange: handlerBatchData
                }}
                action={handlerTableAction}
            />
        </div>
    )
}

export default Supplier
