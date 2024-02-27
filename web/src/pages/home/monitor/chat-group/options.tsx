import { ActionKey, NotifyAppData } from '@/apis/data'
import { ChatGroupItem } from '@/apis/home/monitor/chat-group/types'
import { NotifyApp, Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import dayjs from 'dayjs'

export type ChatGroupTypeCoumnType =
    | ColumnGroupType<ChatGroupItem>
    | ColumnType<ChatGroupItem>

export const columns: ChatGroupTypeCoumnType[] = [
    {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
        width: 150
    },

    {
        title: '所属平台',
        dataIndex: 'app',
        key: 'app',
        width: 150,
        render: (app: NotifyApp) => {
            return NotifyAppData[app]
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 140,
        render: (status: Status) => {
            const { text, color } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    },
    {
        title: 'Hook名称',
        dataIndex: 'hookName',
        key: 'hookName',
        width: 150
    },
    {
        title: 'Hook',
        dataIndex: 'hook',
        key: 'hook',
        width: 300,
        render: (hook: string) => {
            // 替换最后8个字符为********
            return hook.replace(/(.{8})$/, '********')
        }
    },
    {
        title: '描述',
        dataIndex: 'remark',
        key: 'remark'
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

export const tableOperationItems = (
    record: ChatGroupItem
): MenuProps['items'] => [
    {
        key: ActionKey.EDIT,
        label: (
            <Button
                type="link"
                size="small"
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
        type: 'divider'
    },
    {
        key: ActionKey.DELETE,
        label: (
            <Button
                type="link"
                size="small"
                danger
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

export const defaultPadding = 12

export const rightOptions: DataOptionItem[] = []

export const leftOptions: DataOptionItem[] = []

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

export const addChatGroupItems: (DataFormItem | DataFormItem[])[] = [
    [
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
            name: 'app',
            label: '所属平台',
            rules: [
                {
                    required: true,
                    message: '请选择所属平台'
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择所属平台',
                    options: Object.entries(NotifyAppData).map(
                        ([key, value]) => ({
                            label: value,
                            value: Number(key)
                        })
                    )
                }
            }
        }
    ],
    [
        {
            name: 'hookName',
            label: 'Hook名称',
            rules: [
                {
                    required: true,
                    message: '请输入Hook名称'
                }
            ]
        },
        {
            name: 'hook',
            label: 'Hook',
            rules: [
                {
                    required: true,
                    message: '请输入Hook',
                    validator: (_, value, callback) => {
                        // https || http
                        if (!value) {
                            callback('请输入Hook')
                        } else if (!/^(https|http):\/\/.+$/.test(value)) {
                            callback('请输入正确的Hook, https或者http开头')
                        } else {
                            callback()
                        }
                    }
                }
            ]
        }
    ],
    {
        name: 'title',
        label: '告警标题',
        formItemProps: {
            tooltip:
                '告警标题，用于在告警信息中展示, 支持模板语法， 例如: Prometheus告警{{ $tatus }}'
        }
    },
    {
        name: 'template',
        label: '告警内容',
        formItemProps: {
            tooltip: `告警内容，用于在告警信息中展示, 支持模板语法， 例如:
            ## prometheus监控告警【{{ $status }}】

            * 告警时间: {{ $startsAt }}
            * 恢复时间: {{ $endsAt }}
            * 告警标题: {{ $annotations.title }}
            * 告警内容: {{ $annotations.description }}
            * 唯一指纹: {{ $fingerprint }}
            * 告警标识
                * 规则名称: {{ $labels.alertname }}
                * 机器名称: {{ $labels.endpoint }}
                * 实例名称: {{ $labels.instance }}`
        },
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入告警内容',
                showCount: true
            }
        }
    },
    {
        name: 'secret',
        label: 'secret',
        formItemProps: {
            tooltip: 'secret是飞书必传字段'
        }
    },
    {
        name: 'remark',
        label: 'remark',
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入备注',
                showCount: true,
                maxLength: 255
            }
        }
    }
]

export const updateChatGroupItems: (DataFormItem | DataFormItem[])[] = [
    [
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
            name: 'hookName',
            label: 'Hook名称',
            rules: [
                {
                    required: true,
                    message: '请输入Hook名称'
                }
            ]
        }
    ],
    {
        name: 'title',
        label: '告警标题',
        formItemProps: {
            tooltip:
                '告警标题，用于在告警信息中展示, 支持模板语法， 例如: Prometheus告警{{ $tatus }}'
        }
    },
    {
        name: 'template',
        label: '告警内容',
        formItemProps: {
            tooltip: `告警内容，用于在告警信息中展示, 支持模板语法， 例如:
            ## prometheus监控告警【{{ $status }}】

            * 告警时间: {{ $startsAt }}
            * 恢复时间: {{ $endsAt }}
            * 告警标题: {{ $annotations.title }}
            * 告警内容: {{ $annotations.description }}
            * 唯一指纹: {{ $fingerprint }}
            * 告警标识
                * 规则名称: {{ $labels.alertname }}
                * 机器名称: {{ $labels.endpoint }}
                * 实例名称: {{ $labels.instance }}`
        },
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入告警内容',
                showCount: true
            }
        }
    },

    {
        name: 'remark',
        label: 'remark',
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入备注',
                showCount: true,
                maxLength: 255
            }
        }
    }
]
