import { ActionKey, NotifyAppData } from '@/apis/data'
import {
    AlarmGroupItem,
    NotifyMemberItem
} from '@/apis/home/monitor/alarm-group/types'
import { NotifyMember } from '@/apis/home/monitor/alarm-notify/types'
import { ChatGroupSelectItem } from '@/apis/home/monitor/chat-group/types'
import { UserSelectItem } from '@/apis/home/system/user/types'
import { NotifyApp, Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, Checkbox, MenuProps } from 'antd'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import dayjs from 'dayjs'

export const defaultPadding = 12

export const rightOptions: DataOptionItem[] = [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    }
]

export const leftOptions: DataOptionItem[] = []

export type AlarmGroupTableColumnType =
    | ColumnGroupType<AlarmGroupItem>
    | ColumnType<AlarmGroupItem>

export type ChartGroupTableColumnType =
    | ColumnGroupType<ChatGroupSelectItem>
    | ColumnType<ChatGroupSelectItem>

export type MemberTableColumnType =
    | ColumnGroupType<NotifyMemberItem>
    | ColumnType<NotifyMemberItem>

export type NotifyMemberTableColumnType =
    | ColumnGroupType<NotifyMember>
    | ColumnType<NotifyMember>

export const columns: AlarmGroupTableColumnType[] = [
    {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
        width: 200
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        align: 'center',
        width: 120,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    },
    {
        title: '描述信息',
        dataIndex: 'remark',
        key: 'remark',
        width: 400
    },
    {
        title: '创建时间',
        dataIndex: 'createdAt',
        key: 'createdAt',
        width: 200,
        render: (createdAt: string | number) => {
            return dayjs(+createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '更新时间',
        dataIndex: 'updatedAt',
        key: 'updatedAt',
        width: 200,
        render: (updatedAt: string | number) => {
            return dayjs(+updatedAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    }
]

export const chartGroupCoumns: ChartGroupTableColumnType[] = [
    {
        title: '名称',
        dataIndex: 'label',
        key: 'label',
        width: '40%'
    },
    {
        title: '所属APP',
        dataIndex: 'app',
        key: 'app',
        width: '40%',
        align: 'center',
        render: (app: NotifyApp) => {
            return NotifyAppData[app]
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        align: 'center',
        width: '40%',
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    }
]

export const memberCoumns = (
    onChange?: (checked: number, record: NotifyMemberItem) => void
): MemberTableColumnType[] => [
    {
        title: '通知人姓名',
        dataIndex: 'user',
        key: 'user',
        width: '40%',
        render: (user: UserSelectItem) => {
            return user.label
        }
    },

    {
        title: '告警方式',
        dataIndex: 'notifyType',
        key: 'notifyType',
        align: 'center',
        width: '40%',
        render: (notifyType: number, record: NotifyMemberItem) => {
            const email = notifyType & 2
            const sms = notifyType & 4
            const phone = notifyType & 8

            return (
                <span>
                    <Checkbox
                        onChange={(e) =>
                            onChange?.(
                                e.target.checked
                                    ? notifyType | 2
                                    : notifyType & 12,
                                record
                            )
                        }
                        checked={!!onChange ? !!email : undefined}
                        defaultChecked={!!email}
                    >
                        邮箱
                    </Checkbox>
                    <Checkbox
                        onChange={(e) =>
                            onChange?.(
                                e.target.checked
                                    ? notifyType | 4
                                    : notifyType & 10,
                                record
                            )
                        }
                        defaultChecked={!!sms}
                        checked={!!onChange ? !!sms : undefined}
                    >
                        短信
                    </Checkbox>
                    <Checkbox
                        onChange={(e) =>
                            onChange?.(
                                e.target.checked
                                    ? notifyType | 8
                                    : notifyType & 6,
                                record
                            )
                        }
                        defaultChecked={!!phone}
                        checked={!!onChange ? !!phone : undefined}
                    >
                        电话
                    </Checkbox>
                </span>
            )
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        align: 'center',
        width: 120,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    }
]

export const tableOperationItems = (
    record: AlarmGroupItem
): MenuProps['items'] => [
    {
        key: ActionKey.EDIT,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit" />}
            >
                编辑
            </Button>
        )
    },
    record.status === Status.STATUS_DISABLED
        ? {
              key: ActionKey.ENABLE,
              label: (
                  <Button
                      size="small"
                      type="link"
                      icon={<IconFont type="icon-Enable" />}
                  >
                      启用
                  </Button>
              )
          }
        : {
              key: ActionKey.DISABLE,
              label: (
                  <Button
                      size="small"
                      type="link"
                      icon={<IconFont type="icon-disable4" />}
                      danger
                  >
                      禁用
                  </Button>
              )
          },
    {
        key: ActionKey.OPERATION_LOG,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-wj-rz" />}
            >
                操作日志
            </Button>
        )
    },
    {
        type: 'divider'
    },
    {
        key: ActionKey.DELETE,
        label: (
            <Button
                size="small"
                danger
                type="link"
                icon={
                    <IconFont
                        type="icon-shanchu-copy"
                        style={{ color: 'red' }}
                    />
                }
            >
                删除
            </Button>
        )
    }
]

export const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '模糊查询'
    },
    {
        name: 'status',
        label: '状态',
        dataProps: {
            type: 'radio-group',
            parentProps: {
                optionType: 'button',
                defaultValue: Status.STATUS_UNKNOWN,
                options: [
                    {
                        label: '全部',
                        value: Status.STATUS_UNKNOWN
                    },
                    {
                        label: '启用',
                        value: Status.STATUS_ENABLED
                    },
                    {
                        label: '禁用',
                        value: Status.STATUS_DISABLED
                    }
                ]
            }
        }
    }
]

export const editorItems: DataFormItem[] = [
    {
        name: 'name',
        label: '名称',
        rules: [
            {
                required: true,
                message: '请输入名称'
            }
        ]
    },
    {
        name: 'remark',
        label: '描述信息',
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入描述信息',
                showCount: true,
                maxLength: 255
            }
        }
    }
]
