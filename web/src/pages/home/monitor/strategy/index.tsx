import { FC, Key, useContext, useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Form, Space, message } from 'antd'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { CopyOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { ActionKey } from '@/apis/data.tsx'
import {
    StrategyItemType,
    StrategyListRequest,
    defaultStrategyListRequest
} from '@/apis/home/monitor/strategy/types'
import {
    columns,
    leftOptions,
    rightOptions,
    searchItems,
    tableOperationItems
} from './options'
import { Detail } from './child/Detail'
import strategyApi from '@/apis/home/monitor/strategy'
import { ModuleType, Status } from '@/apis/types'
import { BindNotifyObject } from './child/BindNotifyObject'
import { GlobalContext } from '@/context'
import { ImportGroups } from '../strategy-group/child/ImportGroups'
import qs from 'qs'
import PromQLInput from '@/components/Prom/PromQLInput'
import { SysLogDetail } from '../../child/SysLogDetail'
import { BindNotifyTemplate } from './child/BindNotifyTemplate'

const defaultPadding = 12

let fetchTimer: NodeJS.Timeout
const Strategy: FC = () => {
    const { size } = useContext(GlobalContext)
    const navigate = useNavigate()
    const [searchParams, setSearchParams] = useSearchParams()

    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<StrategyItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number | string>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [operateId, setOperateId] = useState<number | undefined>()
    const [actionKey, setActionKey] = useState<ActionKey | undefined>(
        ActionKey.ADD
    )
    const [openBindNotify, setOpenBindNotify] = useState<boolean>(false)

    const [reqParams, setReqParams] = useState<StrategyListRequest>(
        defaultStrategyListRequest
    )

    const [openImportModal, setOpenImportModal] = useState<boolean>(false)
    const [logOpen, setLogOpen] = useState<boolean>(false)
    const [logDataId, setLogDataId] = useState<number | undefined>()
    const [openBindNotifyTemplate, setOpenBindNotifyTemplate] =
        useState<boolean>(false)

    const openLogDetail = (id: number) => {
        setLogOpen(true)
        setLogDataId(id)
    }

    const closeLogDetail = () => {
        setLogOpen(false)
        setLogDataId(undefined)
    }

    const handleOpenImportModal = () => {
        setOpenImportModal(true)
    }

    const handleCloseImportModal = () => {
        setOpenImportModal(false)
    }

    const handleImportOnOk = () => {
        handleCloseImportModal()
        handlerRefresh()
    }

    const handlerOpenDetail = (id?: number) => {
        setOperateId(id)
        setOpenDetail(true)
    }

    const handleOpenBindNotify = (id?: number) => {
        setOperateId(id)
        setOpenBindNotify(true)
    }

    const handleCancelBindNotify = () => {
        setOpenBindNotify(false)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
    }

    const handleOpenBindNotifyTemplate = (id?: number) => {
        setOperateId(id)
        setOpenBindNotifyTemplate(true)
    }

    const handlerCloseBindNotifyTemplate = () => {
        setOpenBindNotifyTemplate(false)
        setOperateId(undefined)
    }

    // 获取数据
    const handlerGetData = () => {
        if (fetchTimer) {
            clearTimeout(fetchTimer)
        }
        fetchTimer = setTimeout(() => {
            setLoading(true)
            strategyApi
                .getStrategyList(reqParams)
                .then((res) => {
                    setDataSource(res.list)
                    setTotal(res.page.total)
                })
                .finally(() => {
                    setLoading(false)
                })
        }, 500)
    }

    // 刷新
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
    }

    // 分页变化
    const handlerTablePageChange = (page: number, pageSize?: number) => {
        setReqParams({
            ...reqParams,
            page: {
                curr: page,
                size: pageSize || reqParams?.page?.size
            }
        })
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: Key[],
        selectedRows: StrategyItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    const toStrategyGroupPage = (record: StrategyItemType) => {
        navigate(`/home/monitor/strategy-group?id=${record.groupId}`)
    }

    const handlerBatchDelete = (ids: number[]) => {
        console.log(ids)
    }

    // 批量修改状态
    const handlebatchChangeStatus = (ids: number[], status: Status) => {
        strategyApi.batchChangeStrategyStatus(ids, status).then(() => {
            handlerRefresh()
        })
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: StrategyItemType) => {
        setActionKey(key)
        switch (key) {
            case ActionKey.STRATEGY_GROUP_LIST:
                toStrategyGroupPage(record)
                break
            case ActionKey.DETAIL:
                handlerOpenDetail(record.id)
                break
            case ActionKey.EDIT:
                handlerOpenDetail(record.id)
                break
            case ActionKey.DELETE:
                handlerBatchDelete([record.id])
                break
            case ActionKey.DISABLE:
                handlebatchChangeStatus([record.id], Status.STATUS_DISABLED)
                break
            case ActionKey.ENABLE:
                handlebatchChangeStatus([record.id], Status.STATUS_ENABLED)
                break
            case ActionKey.STRATEGY_NOTIFY_OBJECT:
                handleOpenBindNotify(record.id)
                break
            case ActionKey.OPERATION_LOG:
                openLogDetail(record.id)
                break
            case ActionKey.STRATEGY_BIND_NOTIFY_TEMPLATE:
                handleOpenBindNotifyTemplate(record.id)
                break
            default:
                break
        }
    }

    const handlerDataOptionAction = (key: ActionKey) => {
        setActionKey(key)
        switch (key) {
            case ActionKey.ADD:
                handlerOpenDetail()
                break
            case ActionKey.BATCH_IMPORT:
                handleOpenImportModal()
                break
            case ActionKey.REFRESH:
                handlerRefresh()
                break
            case ActionKey.RESET:
                setReqParams(defaultStrategyListRequest)
                setSearchParams('')
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (_: any, allValues: any) => {
        setReqParams({
            ...reqParams,
            ...allValues,
            page: defaultStrategyListRequest.page
        })
        handlerRefresh()
    }

    const handleCopyExpr = (expr?: string) => {
        return function () {
            if (expr) {
                navigator.clipboard.writeText(expr)
                message.success('已复制到剪贴板')
            }
        }
    }

    useEffect(() => {
        handlerRefresh()
    }, [reqParams])

    useEffect(() => {
        const searchP = qs.parse(searchParams.toString()) as any
        // 获取prams
        const req: StrategyListRequest = {
            ...reqParams,
            strategyId: +searchP?.strategyId || 0,
            groupId: searchP?.groupId ? +searchP?.groupId : 0
        }
        setReqParams(req)
    }, [searchParams])

    useEffect(() => {
        handlerGetData()
    }, [refresh])

    return (
        <div>
            <BindNotifyTemplate
                title="通知模板"
                strategyId={operateId}
                open={openBindNotifyTemplate}
                width="60%"
                onCancel={handlerCloseBindNotifyTemplate}
                onOk={handlerCloseBindNotifyTemplate}
            />
            <SysLogDetail
                module={ModuleType.ModuleStrategy}
                moduleId={logDataId}
                open={logOpen}
                width={600}
                onClose={closeLogDetail}
            />
            <BindNotifyObject
                open={openBindNotify}
                onClose={handleCancelBindNotify}
                strategyId={operateId}
            />
            <ImportGroups
                width="60%"
                title="批量导入"
                onOk={handleImportOnOk}
                onCancel={handleCloseImportModal}
                open={openImportModal}
            />
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                id={operateId}
                actionKey={actionKey}
                refresh={handlerRefresh}
                disabled={actionKey === ActionKey.DETAIL}
            />
            <div>
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
                    rightOptions={rightOptions(loading)}
                    leftOptions={leftOptions(loading)}
                    action={handlerDataOptionAction}
                />
                <PaddingLine
                    padding={defaultPadding}
                    height={1}
                    borderRadius={4}
                />
            </div>

            <DataTable
                dataSource={dataSource}
                columns={columns(size, {})}
                total={+total}
                loading={loading}
                operationItems={tableOperationItems}
                pageOnChange={handlerTablePageChange}
                rowSelection={{ onChange: handlerBatchData }}
                showIndex={false}
                pageSize={reqParams?.page?.size}
                current={reqParams?.page?.curr}
                action={handlerTableAction}
                expandable={{
                    expandedRowRender: (record: StrategyItemType) => (
                        <Space size="middle" direction="vertical">
                            <Space style={{ width: '100%' }}>
                                <Button
                                    type="primary"
                                    icon={<CopyOutlined />}
                                    size="small"
                                    onClick={handleCopyExpr(record?.expr)}
                                />
                                <Form layout="inline">
                                    <Form.Item>
                                        <PromQLInput
                                            disabled={true}
                                            pathPrefix=""
                                            value={record?.expr}
                                            showBorder={false}
                                        />
                                    </Form.Item>
                                </Form>
                            </Space>
                            {!!record?.remark && (
                                <Space style={{ width: '100%' }}>
                                    <Button
                                        type="text"
                                        size="small"
                                        icon={<InfoCircleOutlined />}
                                    />
                                    <p style={{ margin: 0 }}>
                                        {record?.remark}
                                    </p>
                                </Space>
                            )}
                        </Space>
                    )
                }}
            />
        </div>
    )
}

export default Strategy
