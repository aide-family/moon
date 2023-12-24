import { FC, useEffect, useState } from 'react'
import { Descriptions, Modal } from 'antd'
import { SupplierItemType } from '../type'
import { buildDetailItems } from '../data'

export type DetailProps = {
    open?: boolean
    onClose?: () => void
    supplierId?: string
}

const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, supplierId } = props
    const [supplierDetail, setSupplierDetail] = useState<
        SupplierItemType | undefined
    >()

    useEffect(() => {
        // const res = await getUserDetail(userId)
        // setUserDetail(res)
        setSupplierDetail({
            id: supplierId || '',
            company: 'company_' + supplierId,
            contacts: 'contacts_' + supplierId,
            phone: 'phone_' + supplierId,
            wechat: 'wechat' + supplierId,
            address: 'address_' + supplierId,
            created_at: 0,
            created_by: 'created_by_' + supplierId
        })
    }, [supplierId])

    return (
        <Modal open={open} onCancel={onClose} centered footer={null}>
            <Descriptions
                column={2}
                title="客户详情"
                items={buildDetailItems(supplierDetail)}
            />
        </Modal>
    )
}

export default Detail
