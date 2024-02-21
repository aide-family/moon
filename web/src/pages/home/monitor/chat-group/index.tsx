import {
    ChatGroupItem,
    ListChatGroupRequest,
    defaultListChatGroupRequest
} from '@/apis/home/monitor/chat-group/types'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Form } from 'antd'
import React, { useEffect, useRef, useState } from 'react'
import {
    columns,
    defaultPadding,
    leftOptions,
    rightOptions,
    searchItems,
    tableOperationItems
} from './options'
import { ActionKey } from '@/apis/data'
import chatGroupApi from '@/apis/home/monitor/chat-group'
import EditChatGroupModal from './child/EditChatGroupModal'

export interface ChatGroupProps {}

let timer: NodeJS.Timeout | null = null
const ChatGroup: React.FC<ChatGroupProps> = () => {
    const [queryForm] = Form.useForm()
    const operationRef = useRef<HTMLDivElement>(null)
    const [searchRequest, setSearchRequest] = useState<ListChatGroupRequest>(
        defaultListChatGroupRequest
    )
    const [refresh, setRefresh] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(false)
    const [dataSource, setDataSource] = useState<ChatGroupItem[]>([])
    const [total, setTotal] = useState<number>(0)
    const [opChatGroupId, setOpChatGroupId] = useState<number>()
    const [openChatGroupModal, setChatGroupModal] = useState<boolean>(false)

    const handleRefresh = () => {
        setRefresh((p) => !p)
    }

    const handleOpenChatGroupModal = (id?: number) => {
        setOpChatGroupId(id)
        setChatGroupModal(true)
    }

    const handleChatGroupModalClose = () => {
        setChatGroupModal(false)
        setOpChatGroupId(undefined)
    }

    const handleChatGroupModalOk = () => {
        handleChatGroupModalClose()
        handleRefresh()
    }

    const handleGetChatGroupList = () => {
        if (timer) {
            clearTimeout(timer)
        }
        timer = setTimeout(() => {
            setLoading(true)
            chatGroupApi
                .getChatGroupList(searchRequest)
                .then((data) => {
                    setDataSource(data?.list || [])
                    setTotal(data?.page.total || 0)
                })
                .finally(() => {
                    setLoading(false)
                })
        }, 500)
    }

    const handlerDataOptionAction = (key: ActionKey) => {
        switch (key) {
            case ActionKey.REFRESH:
                handleRefresh()
                break
            case ActionKey.RESET:
                setSearchRequest(defaultListChatGroupRequest)
                break
            case ActionKey.EXPORT:
                break
            case ActionKey.BATCH_IMPORT:
                break
            case ActionKey.BATCH_EXPORT:
                break
            case ActionKey.ADD:
                handleOpenChatGroupModal()
                break
            default:
                break
        }
    }

    const handlerTablePageChange = (page: number, size: number) => {
        setSearchRequest({
            ...searchRequest,
            page: {
                curr: page,
                size: size
            }
        })
    }

    const handlerTableAction = (key: ActionKey, item: ChatGroupItem) => {
        switch (key) {
            case ActionKey.EDIT:
                handleOpenChatGroupModal(item.id)
                break
            case ActionKey.DELETE:
                break
            case ActionKey.ENABLE:
                // handleChangeStatus([item.id], Status.STATUS_ENABLED)
                break
            case ActionKey.DISABLE:
                // handleChangeStatus([item.id], Status.STATUS_DISABLED)
                break
            default:
                break
        }
    }

    const handlerSearchFormValuesChange = (_: any, allValues: any) => {
        setSearchRequest({
            ...searchRequest,
            ...allValues
        })
    }

    useEffect(() => {
        handleRefresh()
    }, [searchRequest])

    useEffect(() => {
        handleGetChatGroupList()
    }, [refresh])

    return (
        <div>
            <EditChatGroupModal
                chatGroupId={opChatGroupId}
                open={openChatGroupModal}
                onClose={handleChatGroupModalClose}
                onOk={handleChatGroupModalOk}
            />
            <div ref={operationRef}>
                <RouteBreadcrumb />
                <HeightLine />
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearchFormValuesChange
                    }}
                />
                <HeightLine />
                <DataOption
                    queryForm={queryForm}
                    rightOptions={rightOptions}
                    leftOptions={leftOptions}
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
                columns={columns}
                operationRef={operationRef}
                total={+total}
                loading={loading}
                operationItems={tableOperationItems}
                pageOnChange={handlerTablePageChange}
                action={handlerTableAction}
                pageSize={searchRequest?.page?.size}
                current={searchRequest?.page?.curr}
            />
        </div>
    )
}

export default ChatGroup
