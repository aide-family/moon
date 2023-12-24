import { useEffect, useRef, useState, FC, Key } from 'react'

import type { ColumnType, ColumnGroupType } from 'antd/es/table'
import type { DataFormItem } from '@/components/Data'
import type { CustomerItemType } from './type'

import { Button, Table } from 'antd'
import { useForm } from 'antd/es/form/Form'
import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { operationItems } from '@/components/Data/DataOption/option'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import Detail from './child/Detail'
import { ActionKey } from '@/apis/data'

const defaultPadding = 12

let dataSourceTmp: CustomerItemType[] = []
for (let i = 0; i < 100; i++) {
    dataSourceTmp.push({
        id: i + 1 + '',
        company: 'company_' + i,
        contacts: 'contacts_' + i,
        phone: 'phone_' + i,
        chat_group: 'chat_group_' + i,
        address: 'address_' + i
    })
}

const Customer: FC = () => {
    const [queryForm] = useForm()
    const oprationRef = useRef<HTMLDivElement>(null)
    const searchItems: DataFormItem[] = [
        {
            label: '公司名称',
            name: 'company'
        },
        {
            label: '客户名称',
            name: 'contacts'
        },
        {
            label: '联系电话',
            name: 'phone'
        }
    ]

    const columns: (
        | ColumnGroupType<CustomerItemType>
        | ColumnType<CustomerItemType>
    )[] = [
        Table.SELECTION_COLUMN,
        {
            title: '公司名称',
            dataIndex: 'company',
            key: 'company',
            width: 220,
            ellipsis: true
        },
        {
            title: '客户名称',
            dataIndex: 'contacts',
            width: 200,
            ellipsis: true
        },
        {
            title: '联系电话',
            dataIndex: 'phone',
            width: 200
        },
        {
            title: '微信群',
            dataIndex: 'chat_group',
            width: 200
        },
        {
            title: '公司地址',
            dataIndex: 'address'
        }
    ]
    const [dataSource, setDataSource] = useState<CustomerItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [customerDetail, setCustomerDetail] = useState<
        CustomerItemType | undefined
    >()

    const handlerOpenDetail = () => {
        setOpenDetail(true)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
        setCustomerDetail(undefined)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: CustomerItemType) => {
        console.log(key, record)
        switch (key) {
            case ActionKey.DETAIL:
                handlerOpenDetail()
                setCustomerDetail(record)
                break
        }
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: Key[],
        selectedRows: CustomerItemType[]
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
            setDataSource(dataSourceTmp)
            setTotal(203)
            setLoading(false)
        }, 1500)
    }, [refresh])

    return (
        <div className="bodyContent">
            <Detail
                customerId={customerDetail?.id}
                open={openDetail}
                onClose={handlerCloseDetail}
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

export default Customer
