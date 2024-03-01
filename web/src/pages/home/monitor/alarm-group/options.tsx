import { ActionKey } from '@/apis/data'
import { AlarmGroupItem } from '@/apis/home/monitor/alarm-group/types'
import { ChatGroupSelectItem } from '@/apis/home/monitor/chat-group/types'
import { Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { ColumnGroupType, ColumnType } from 'antd/es/table'

export const defaultPadding = 12

export const rightOptions: DataOptionItem[] = []

export const leftOptions: DataOptionItem[] = []

export type AlarmGroupTableColumnType =
    | ColumnGroupType<AlarmGroupItem>
    | ColumnType<AlarmGroupItem>

export type ChartGroupTableColumnType =
    | ColumnGroupType<ChatGroupSelectItem>
    | ColumnType<ChatGroupSelectItem>

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
        width: 140,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    },
    {
        title: '描述信息',
        dataIndex: 'remark',
        key: 'remark'
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
