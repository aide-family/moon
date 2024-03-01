import { ActionKey } from '@/apis/data'
import {
    AlarmPageItem,
    ListAlarmPageRequest
} from '@/apis/home/monitor/alarm-page/types'
import { PageReqType, Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { ColumnGroupType } from 'antd/es/table'
import { ColumnType } from 'antd/lib/table'
import dayjs from 'dayjs'

export const defaultPadding = 12

export const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '告警页面'
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

export const rightOptions: DataOptionItem[] = [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    }
]

export const leftOptions: DataOptionItem[] = [
    {
        key: ActionKey.BATCH_IMPORT,
        label: <Button type="primary">批量导入</Button>
    }
]

export type AlarmPageTableColumnType =
    | ColumnGroupType<AlarmPageItem>
    | ColumnType<AlarmPageItem>

export const columns: AlarmPageTableColumnType[] = [
    {
        title: '告警页面名称',
        dataIndex: 'name',
        key: 'name',
        width: 200,
        fixed: 'left'
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 100,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    },
    {
        title: '图标',
        dataIndex: 'icon',
        key: 'icon',
        width: 100,
        render: (icon: string, item: AlarmPageItem) => {
            return icon ? (
                <Button
                    type="link"
                    icon={
                        <IconFont type={icon} style={{ color: item.color }} />
                    }
                />
            ) : (
                '-'
            )
        }
    },
    {
        title: '页面识别颜色',
        dataIndex: 'color',
        key: 'color',
        width: 120,
        align: 'center',
        render: (color: string) => {
            return (
                <div
                    color={color}
                    style={{
                        backgroundColor: color,
                        color: '#fff',
                        width: '80%',
                        textAlign: 'center'
                    }}
                >
                    {color}
                </div>
            )
        }
    },
    {
        title: '备注',
        dataIndex: 'remark',
        key: 'remark',
        render: (remark: string) => {
            return remark || '-'
        }
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
    record: AlarmPageItem
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
    { type: 'divider' },
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

export const defaultPageReqest: PageReqType = {
    size: 10,
    curr: 1
}

export const defaultListAlarmPageRequest: ListAlarmPageRequest = {
    page: defaultPageReqest,
    keyword: '',
    status: 0
}

export type alarmPageDataFormType = {
    name: string
    icon?: string
    color?: string
    remark?: string
}

export const alarmPageDataFormItems: (DataFormItem[] | DataFormItem)[] = [
    {
        label: '告警页面名称',
        name: 'name',
        rules: [
            {
                required: true,
                message: '请输入告警页面名称'
            }
        ]
    },
    [
        {
            label: '图标',
            name: 'icon',
            rules: []
        },
        {
            label: '页面识别颜色',
            name: 'color',
            rules: [],
            dataProps: {
                type: 'color',
                parentProps: {
                    showText: true,
                    defaultFormat: 'hex'
                }
            }
        }
    ],
    {
        label: '说明信息',
        name: 'remark',
        rules: [],
        dataProps: {
            type: 'textarea',
            parentProps: {
                showCount: true,
                maxLength: 200,
                placeholder: '请输入至多200个字符的备注信息'
            }
        }
    }
]
