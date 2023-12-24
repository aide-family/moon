import { DescriptionsProps } from 'antd'
import dayjs from 'dayjs'
import { CustomerItemType } from '../type'

export const buildDetailItems = (
    customerDetail: CustomerItemType
): DescriptionsProps['items'] => {
    const { chat_group, company, contacts, created_at, created_by, phone } =
        customerDetail
    return [
        {
            key: '3',
            label: '公司名称',
            children: company
        },
        {
            key: '4',
            label: '联系人',
            children: contacts
        },
        {
            key: '5',
            label: '联系电话',
            children: phone
        },
        {
            key: '6',
            label: '微信群',
            children: chat_group
        },

        {
            key: '8',
            label: '创建时间',
            children: created_at
                ? dayjs.unix(created_at).format('YYYY-MM-DD HH:mm:ss')
                : '-'
        },
        {
            key: '9',
            label: '创建人',
            children: created_by
        },
        {
            key: '7',
            label: '公司地址',
            children: contacts
        }
    ]
}
