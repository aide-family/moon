import { ActionKey } from '@/apis/data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { SettingOutlined } from '@ant-design/icons'
import { Button, Space, Switch } from 'antd'
import { AddChartButton } from './child/AddChartButton'
import { DataFormItem } from '@/components/Data'
import { AddDashboardButton } from './child/AddDashboardModal'
import dashboardApi from '@/apis/home/dashboard'
import { defaultListChartRequest } from '@/apis/home/dashboard/types'
import { DefaultOptionType } from 'antd/es/select'

export const rightOptions = (autoRefresh?: boolean): DataOptionItem[] => [
    {
        label: (
            <Space>
                自动刷新:
                <Switch
                    checked={autoRefresh}
                    checkedChildren="开"
                    unCheckedChildren="关"
                />
            </Space>
        ),
        key: ActionKey.AUTO_REFRESH
    },
    {
        label: <Button type="primary">刷新</Button>,
        key: ActionKey.REFRESH
    },
    {
        label: <Button type="primary" icon={<SettingOutlined />} />,
        key: ActionKey.CONFIG_DASHBOARD_CHART
    }
]

export const leftOptions = (refresh?: () => void): DataOptionItem[] => [
    {
        label: <AddChartButton />,
        key: ActionKey.ADD
    },
    {
        label: <AddDashboardButton refresh={refresh} />,
        key: ActionKey.ADD
    }
]

const getChartSelect = (keyword: string): Promise<DefaultOptionType[]> => {
    return dashboardApi
        .getChartList({ ...defaultListChartRequest, keyword })
        .then(({ list }) => {
            if (!list || list.length === 0) return []
            return list.map(({ id, title }) => ({ label: title, value: id }))
        })
}

export const addDashboardOptions: DataFormItem[] = [
    {
        label: '图表名称',
        name: 'title',
        rules: [
            {
                required: true,
                message: '请输入图表名称'
            }
        ]
    },
    {
        label: '颜色',
        name: 'color',
        rules: [
            {
                required: true,
                message: '请输入图表名称'
            }
        ],
        dataProps: {
            type: 'color',
            parentProps: {
                showText: true,
                defaultFormat: 'hex'
            }
        }
    },
    {
        label: '大盘说明',
        name: 'remark',
        dataProps: {
            type: 'textarea',
            parentProps: {
                showCount: true,
                maxLength: 255,
                placeholder: '请输入大盘说明'
            }
        }
    },
    // chartIds
    {
        label: '绑定图表',
        name: 'chartIds',
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                handleFetch: getChartSelect,
                defaultOptions: [],
                selectProps: {
                    mode: 'multiple',
                    placeholder: '请选择图表'
                }
            }
        }
    }
]

export const addChartOptions: DataFormItem[] = [
    {
        label: '图表名称',
        name: 'title',
        rules: [
            {
                required: true,
                message: '请输入图表名称'
            }
        ]
    },
    {
        label: '图表链接',
        name: 'url',
        rules: [
            {
                required: true,
                message: '请输入图表链接'
            },
            {
                validator(_, value, callback) {
                    // http or https
                    if (!value) {
                        callback('请输入图表链接')
                    } else if (
                        !/^(http|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?$/.test(
                            value
                        )
                    ) {
                        callback('请输入正确的链接, http或https开始')
                    } else {
                        callback()
                    }
                }
            }
        ]
    },
    {
        label: '图表说明',
        name: 'remark',
        dataProps: {
            type: 'textarea',
            parentProps: {
                showCount: true,
                maxLength: 255,
                placeholder: '请输入图表说明'
            }
        }
    }
]
