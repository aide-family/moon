import { useEffect, useState, FC } from 'react'
import { Descriptions, Modal } from 'antd'
import dayjs from 'dayjs'
import { CustomerItemType } from '../type'
import { buildDetailItems } from './options'

export type DetailProps = {
    open?: boolean
    onClose?: () => void
    customerId?: string
}

const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, customerId } = props
    const [customerDetail, setCustomerDetail] = useState<CustomerItemType>({
        id: customerId || ''
    })

    useEffect(() => {
        // const res = await getUserDetail(userId)
        // setUserDetail(res)
        setCustomerDetail({
            id: customerId || '',
            company: 'company_' + customerId,
            contacts: 'contacts_' + customerId,
            phone: 'phone_' + customerId,
            chat_group: 'chat_group_' + customerId,
            address: 'address_' + customerId,
            created_at: dayjs().unix(),
            created_by: 'created_by_' + customerId
        })
    }, [customerId])

    return (
        <Modal open={open} onCancel={onClose} centered footer={null}>
            <Descriptions
                column={2}
                title="客户详情"
                items={buildDetailItems(customerDetail)}
            />
        </Modal>
    )
}

export default Detail
