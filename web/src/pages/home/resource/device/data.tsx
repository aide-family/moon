import { DeviceItemType } from './type'

export const defaultData: DeviceItemType[] = [
    {
        host_name: '主机1',
        type: {
            name: '设备类型1',
            icon: 'icon-38'
        },
        status: {
            name: '状态1',
            color: 'green'
        },
        sn: '序列号1',
        supplier: {
            name: '供应商1'
        },
        node: {
            name: '节点1'
        },
        manage_ip: '192.168.1.1',
        ipmi: '192.168.1.1',
        manage_port: '22',
        source: {
            name: '来源1'
        },
        remark: '备注1+++++++++' + Array(100).fill('1').join(''),
        deleted: 0,
        created_at: 1700000000000,
        updated_at: 1700000000000,
        created_by: '创建人1',
        updated_by: '更新人1',
        id: '124214',
        space_instance: '空间实例1'
    }
]
