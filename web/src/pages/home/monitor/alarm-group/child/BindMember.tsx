import { Button, Form, Space, Table } from 'antd'
import { DefaultOptionType } from 'antd/es/select'
import React, { useState } from 'react'
import { MemberTableColumnType, memberCoumns } from '../options'
import { IconFont } from '@/components/IconFont/IconFont'
import userApi from '@/apis/home/system/user'
import FetchSelect from '@/components/Data/FetchSelect'
import EditUserModal from '@/pages/home/system/user/child/EditModal'
import { NotifyMemberItem } from '@/apis/home/monitor/alarm-group/types'
import { UserSelectItem } from '@/apis/home/system/user/types'

export interface BindMemberProps {
    value?: NotifyMemberItem[]
    defaultValue?: NotifyMemberItem[]
    onChange?: (value?: NotifyMemberItem[]) => void
    disabled?: boolean
}

export const BindMember: React.FC<BindMemberProps> = (props) => {
    const { value, defaultValue, onChange, disabled } = props
    const [memberSelectFrom] = Form.useForm<{ groups: DefaultOptionType[] }>()
    const [editUserModalVisible, setEditUserModalVisible] =
        useState<boolean>(false)
    const removeMember = (memberItem: NotifyMemberItem) => {
        if (memberItem) {
            const membersTmp = value?.filter(
                (item) => item.memberId !== memberItem.memberId
            )
            onChange?.(membersTmp)
        }
    }

    const appendMember = () => {
        memberSelectFrom.validateFields().then((values) => {
            if (!value) {
                return
            }
            const membersTmp = value
            values.groups.map((item) => {
                membersTmp.push({
                    memberId: item.value as number,
                    user: item.title as UserSelectItem,
                    notifyType: 0,
                    status: 0,
                    id: 0
                })
            })

            // 去重
            onChange?.(
                value?.filter((item, index, arr) => {
                    return (
                        arr.findIndex(
                            (itemTmp) => itemTmp.memberId === item.memberId
                        ) === index
                    )
                })
            )
            memberSelectFrom.resetFields()
        })
    }

    const notifyTypeOnChange = (checked: number, record: NotifyMemberItem) => {
        const valueTmp = value?.map((item) => {
            if (item.memberId === record.memberId) {
                item.notifyType = checked
            }
            return item
        })
        onChange?.(valueTmp)
    }

    const memberCoumnsOptions: MemberTableColumnType[] = [
        ...memberCoumns(notifyTypeOnChange),
        {
            title: '操作',
            dataIndex: 'action',
            key: 'action',
            align: 'center',
            width: 80,
            render: (_: any, record: NotifyMemberItem) => {
                return (
                    <Button
                        size="small"
                        danger
                        type="link"
                        onClick={() => removeMember(record)}
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

    const getMemberList = (keyword?: string): Promise<DefaultOptionType[]> => {
        return userApi
            .userSelect({
                page: { curr: 1, size: 10 },
                keyword
            })
            .then((data) => {
                if (!data || !data.list || data.list.length === 0) return []
                return data.list?.map((item) => {
                    const option: DefaultOptionType = {
                        label: item.label,
                        title: item,
                        value: item.value
                    }
                    return option
                })
            })
    }
    return (
        <>
            <EditUserModal
                open={editUserModalVisible}
                onClose={() => setEditUserModalVisible(false)}
                onOk={() => setEditUserModalVisible(false)}
            />
            <Form
                layout="inline"
                form={memberSelectFrom}
                onFinish={appendMember}
            >
                <Form.Item name="groups" label={<b>通知人</b>}>
                    {!disabled && (
                        <FetchSelect
                            handleFetch={getMemberList}
                            width={400}
                            defaultOptions={[]}
                            selectProps={{
                                placeholder: '请选择通知人',
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
                            onClick={() => setEditUserModalVisible(true)}
                        >
                            去创建通知人
                        </Button>
                    </Space>
                )}
            </Form>
            {value && value.length > 0 && (
                <Table
                    columns={memberCoumnsOptions}
                    dataSource={value || defaultValue}
                    size="small"
                    pagination={false}
                    rowKey={(record) => record.id}
                />
            )}
        </>
    )
}
