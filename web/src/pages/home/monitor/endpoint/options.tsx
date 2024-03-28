import { DataOptionItem } from '@/components/Data/DataOption/DataOption.tsx'
import { ActionKey } from '@/apis/data.tsx'
import { Badge, Button, MenuProps } from 'antd'
import { DataFormItem } from '@/components/Data'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import { PrometheusServerItem } from '@/apis/home/monitor/endpoint/types'
import dayjs from 'dayjs'
import { Status, StatusMap } from '@/apis/types'
import { IconFont } from '@/components/IconFont/IconFont'

export const defaultPadding = 12

export const leftOptions: DataOptionItem[] = [
    // {
    //     key: ActionKey.BATCH_IMPORT,
    //     label: (
    //         <Button type="primary" loading={loading}>
    //             批量导入
    //         </Button>
    //     )
    // }
]

export const rightOptions = (loading?: boolean): DataOptionItem[] => [
    {
        key: ActionKey.REFRESH,
        label: (
            <Button type="primary" loading={loading}>
                刷新
            </Button>
        )
    }
]

export const operationItems = (
    record: PrometheusServerItem
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
        label: '关键词'
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

export type EndpointColumnType =
    | ColumnType<PrometheusServerItem>
    | ColumnGroupType<PrometheusServerItem>

export const columns: EndpointColumnType[] = [
    {
        title: '数据源名称',
        dataIndex: 'name',
        key: 'name',
        width: 220
    },
    {
        title: '端点',
        dataIndex: 'endpoint',
        key: 'endpoint',
        width: 220
    },
    {
        title: '端点状态',
        dataIndex: 'status',
        key: 'status',
        width: 140,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    },
    {
        title: '备注',
        dataIndex: 'remark',
        key: 'remark',
        ellipsis: true
    },
    {
        title: '创建时间',
        dataIndex: 'createdAt',
        key: 'createdAt',
        width: 220,
        render: (createdAt: number | string) => {
            return dayjs(+createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '更新时间',
        dataIndex: 'updatedAt',
        key: 'updatedAt',
        width: 220,
        render: (updatedAt: number | string) => {
            return dayjs(+updatedAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    }
]

export const editModalFormItems: DataFormItem[] = [
    {
        name: 'name',
        label: '数据源名称',
        rules: [
            {
                required: true,
                message: '请输入数据源名称'
            }
        ]
    },
    {
        name: 'endpoint',
        label: '端点',
        rules: [
            {
                required: true,
                message: '请输入端点',
                validator: (_, value, callback) => {
                    // http或者https
                    if (!value) {
                        callback('请输入端点')
                    } else if (!/^(http|https):\/\/.+/.test(value)) {
                        callback('请输入正确的端点')
                    } else {
                        callback()
                    }
                }
            }
        ]
    },
    {
        name: 'remark',
        label: '备注',
        dataProps: {
            type: 'textarea',
            parentProps: {
                showCount: true,
                maxLength: 255,
                placeholder: '请输入备注'
            }
        }
    }
]
