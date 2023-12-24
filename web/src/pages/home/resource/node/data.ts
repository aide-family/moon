import { NodeItemType } from './type'

export const defaultNodes: NodeItemType[] = [
    {
        created_by: 'admin',
        created_at: 1600000000,
        updated_by: 'admin',
        updated_at: 1600000000,
        id: '1',
        space_instance: '1',
        name: 'node1',
        cname: 'node1',
        band_width: 100,
        status: {
            id: '1',
            name: '正常',
            color: '#67C23A'
        },
        supplier_chat: '1',
        customer_chat: '1',
        type: {
            id: '1',
            name: '物理机'
        },
        purpose: {
            id: '1',
            name: '生产'
        },
        is_monitor: 1,
        is_jump: 1,
        customer: {
            id: '1',
            name: '客户1'
        },
        ikuai: '1',
        remark: '备注' + Array(100).fill('1备注').join('-'),
        resource_supplier: {
            id: '1',
            name: '供应商1'
        }
    }
]
