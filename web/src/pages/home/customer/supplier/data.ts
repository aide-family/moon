import type { DescriptionsProps } from 'antd'
import type { SupplierItemType } from './type'
import dayjs from 'dayjs'

export const defaultData: SupplierItemType[] = [
    {
        id: '1',
        company: '公司名称',
        contacts: '联系人',
        phone: '联系电话',
        wechat: '微信群',
        address: '公司地址',
        created_at: 1629782400,
        created_by: '创建人',
        type: {
            label: '客户类型',
            value: 0
        },
        updated_at: 1629782400,
        updated_by: '更新人'
    }
]

export const buildDetailItems = (
    supplierDetail?: SupplierItemType
): DescriptionsProps['items'] => {
    if (!supplierDetail) {
        return []
    }
    const { company, contacts, phone, wechat, address, created_at } =
        supplierDetail
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
            children: wechat
        },
        {
            key: '7',
            label: '公司地址',
            children: address
        },
        {
            key: '8',
            label: '创建时间',
            children: created_at
                ? dayjs.unix(created_at).format('YYYY-MM-DD HH:mm:ss')
                : '-'
        }
    ]
}
