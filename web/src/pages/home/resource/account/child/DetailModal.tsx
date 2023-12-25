import { GlobalContext } from '@/context'
import { FC, useContext, useEffect, useState } from 'react'
import { AccountItemType } from '../type'
import { Descriptions, Modal, Spin, Watermark, message } from 'antd'
import { buildDetailItems } from '../options'
import { GetAccountDetail } from '@/apis/home/resource/account'

export type DetailModalProps = {
    open: boolean
    onClose: () => void
    id?: string
}

const DetailModal: FC<DetailModalProps> = (props) => {
    const { user } = useContext(GlobalContext)
    const { open, onClose, id } = props
    const [data, setData] = useState<AccountItemType | undefined>()
    const [loading, setLoading] = useState<boolean>(false)

    const fetchDetail = () => {
        setLoading(true)
        GetAccountDetail(id || '', {
            ERROR: (msg) => {
                message.error(msg)
            }
        })
            .then((res) => {
                setData(res)
            })
            .finally(() => setLoading(false))
    }

    useEffect(() => {
        if (open) {
            fetchDetail()
        } else {
            setData(undefined)
        }
    }, [open])

    return (
        <Modal
            title="设备详情"
            open={open}
            onCancel={onClose}
            footer={null}
            width="50vw"
        >
            <Spin spinning={loading}>
                <Watermark content={user?.username} className="wh100">
                    <Descriptions
                        bordered
                        column={2}
                        items={buildDetailItems(data)}
                    />
                </Watermark>
            </Spin>
        </Modal>
    )
}

export default DetailModal
