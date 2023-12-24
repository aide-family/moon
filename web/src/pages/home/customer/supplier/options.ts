import type { DataFormItem } from '@/components/Data'

export const searchItems: DataFormItem[] = [
    {
        name: 'company',
        label: '公司名称'
    },
    {
        name: 'contacts',
        label: '客户名称'
    },
    {
        name: 'type',
        label: '客户类型',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择客户类型',
                options: [
                    {
                        label: '全部',
                        value: 0
                    }
                ]
            }
        }
    }
]
