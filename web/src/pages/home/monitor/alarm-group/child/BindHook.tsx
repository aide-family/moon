import {
    ChatGroupSelectItem,
    defaultSelectChatGroupReques
} from '@/apis/home/monitor/chat-group/types'
import { NotifyApp, Status } from '@/apis/types'
import { Button, Form, Space, Table } from 'antd'
import { DefaultOptionType } from 'antd/es/select'
import React, { useState } from 'react'
import { ChartGroupTableColumnType, chartGroupCoumns } from '../options'
import { IconFont } from '@/components/IconFont/IconFont'
import FetchSelect from '@/components/Data/FetchSelect'
import chatGroupApi from '@/apis/home/monitor/chat-group'
import EditChatGroupModal from '../../chat-hook/child/EditChatGroupModal'
import {ActionKey} from "@/apis/data.tsx";

export interface BindHookProps {
    value?: ChatGroupSelectItem[]
    defaultValue?: ChatGroupSelectItem[]
    onChange?: (value?: ChatGroupSelectItem[]) => void
    disabled?: boolean
}

export const BindHook: React.FC<BindHookProps> = (props) => {
    const { value, defaultValue, onChange, disabled } = props

    const [chatSelectFrom] = Form.useForm<{ groups: DefaultOptionType[] }>()
    const [editChatGroupModalVisible, setEditChatGroupModalVisible] =
        useState<boolean>(false)
    const removeDetailChatGroup = (chatItem: ChatGroupSelectItem) => {
        if (chatItem) {
            const chatGroupsTmp = value?.filter(
                (item) => item.value !== chatItem.value
            )
            onChange?.(chatGroupsTmp)
        }
    }

    const appendChatGroup = () => {
        chatSelectFrom.validateFields().then((values) => {
            if (!value) {
                return
            }
            const chatGroupsTmp = value
            values.groups.map((item) => {
                chatGroupsTmp.push({
                    value: item.value as number,
                    label: item.label as string,
                    app: item.title
                        ? (+item.title as NotifyApp)
                        : NotifyApp.NOTIFY_APP_CUSTOM,
                    status: Status.STATUS_ENABLED
                })
            })

            // 去重
            onChange?.(
                chatGroupsTmp?.filter((item, index, arr) => {
                    return (
                        arr.findIndex(
                            (itemTmp) => itemTmp.value === item.value
                        ) === index
                    )
                })
            )
            chatSelectFrom.resetFields()
        })
    }
    const chartGroupCoumnsOptions: ChartGroupTableColumnType[] = [
        ...chartGroupCoumns,
        {
            title: '操作',
            key: 'action',
            align: 'center',
            width: 80,
            render: (_: any, record: ChatGroupSelectItem) => {
                return (
                    <Button
                        size="small"
                        danger
                        type="link"
                        onClick={() => removeDetailChatGroup(record)}
                        disabled={disabled}
                        icon={
                            <IconFont
                                disabled={disabled}
                                type="icon-shanchu-copy"
                            />
                        }
                    >
                        删除
                    </Button>
                )
            }
        }
    ]

    const getChatGroupList = (
        keyword?: string
    ): Promise<DefaultOptionType[]> => {
        return chatGroupApi
            .getChatGroupSelect({
                ...defaultSelectChatGroupReques,
                keyword
            })
            .then((data) => {
                if (!data || !data.list || data.list.length === 0) return []
                return data.list?.map((item) => {
                    const option: DefaultOptionType = {
                        label: item.label,
                        app: item.app,
                        status: item.status,
                        value: item.value,
                        title: item.app + ''
                    }
                    return option
                })
            })
    }
    return (
        <>
            <EditChatGroupModal
                open={editChatGroupModalVisible}
                onClose={() => setEditChatGroupModalVisible(false)}
                onOk={() => setEditChatGroupModalVisible(false)}
                action={ActionKey.ADD}
            />
            <Form
                layout="inline"
                form={chatSelectFrom}
                onFinish={appendChatGroup}
            >
                <Form.Item name="groups" label={<b>机器人</b>}>
                    {!disabled && (
                        <FetchSelect
                            handleFetch={getChatGroupList}
                            width={400}
                            defaultOptions={[]}
                            selectProps={{
                                placeholder: '请选择机器人',
                                mode: 'multiple',
                                labelInValue: true,
                                disabled: disabled
                            }}
                        />
                    )}
                </Form.Item>
                {!disabled && (
                    <Space size={8}>
                        <Button
                            type="primary"
                            htmlType="submit"
                            disabled={disabled}
                        >
                            添加
                        </Button>
                        <Button
                            type="link"
                            disabled={disabled}
                            onClick={() => setEditChatGroupModalVisible(true)}
                        >
                            去创建机器人
                        </Button>
                    </Space>
                )}
            </Form>
            {value && value.length > 0 && (
                <Table
                    columns={chartGroupCoumnsOptions}
                    dataSource={value || defaultValue}
                    size="small"
                    pagination={false}
                    rowKey={(record) => record.value}
                />
            )}
        </>
    )
}
