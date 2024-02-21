import { FC, useContext, useEffect, useState } from 'react'
import { Button, Modal, Spin, message } from 'antd'
import SyntaxHighlighter from 'react-syntax-highlighter'
import {
    atomOneDark,
    atomOneLight
} from 'react-syntax-highlighter/dist/esm/styles/hljs'
import { StrategyGroupItemType } from '@/apis/home/monitor/strategy-group/types.ts'
import yaml from 'js-yaml'
import strategyGroupApi from '@/apis/home/monitor/strategy-group'
import { GlobalContext } from '@/context'
import { Map } from '@/apis/types'
import { CopyOutlined } from '@ant-design/icons'

export interface DetailProps {
    open: boolean
    onClose: () => void
    id?: number
}

type StrategyInfoType = {
    alert: string
    expr: string
    for: string
    labels?: Map
    annotations?: Map
}

type strategyGroupInfoType = {
    name: string
    rules: StrategyInfoType[]
}

const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, id } = props
    const { sysTheme } = useContext(GlobalContext)

    const [detail, setDetail] = useState<StrategyGroupItemType>()
    const [strategyGroupInfo, setStrategyGroupInfo] = useState<{
        groups: strategyGroupInfoType[]
    }>()
    const [loading, setLoading] = useState<boolean>(false)

    const buildStrategyGroupInfo = (data: StrategyGroupItemType) => {
        if (!data) return
        const groups: strategyGroupInfoType[] = []
        const group: strategyGroupInfoType = {
            name: data.name,
            rules: []
        }
        data?.strategies?.forEach((item) => {
            const strategyInfo: StrategyInfoType = {
                alert: item.alert,
                expr: item.expr,
                for: `${item.duration?.value}${item.duration?.unit}`,
                labels: item.labels,
                annotations: item.annotations
            }
            group.rules.push(strategyInfo)
        })
        groups.push(group)
        setStrategyGroupInfo({ groups })
    }

    const fetchDetail = () => {
        if (!id) return
        setLoading(true)
        strategyGroupApi
            .getStrategyGroupDetail({ id: id })
            .then(({ detail }) => {
                setDetail(detail)
                buildStrategyGroupInfo(detail)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const handleCopy = () => {
        const text = yaml.dump(strategyGroupInfo)
        navigator.clipboard.writeText(text)
        message.success('复制成功')
    }

    useEffect(() => {
        if (!open) return
        fetchDetail()
    }, [id])

    return (
        <Modal
            title={`${detail?.name} 详情`}
            open={open}
            onCancel={onClose}
            onOk={onClose}
            width="70%"
            destroyOnClose={true}
        >
            {loading ? (
                <Spin />
            ) : (
                <div style={{ position: 'relative' }}>
                    <Button
                        type="link"
                        icon={<CopyOutlined />}
                        onClick={handleCopy}
                        style={{ position: 'absolute', top: 8, right: 8 }}
                    />
                    <SyntaxHighlighter
                        language="yaml"
                        style={sysTheme === 'dark' ? atomOneDark : atomOneLight}
                    >
                        {yaml.dump(strategyGroupInfo)}
                    </SyntaxHighlighter>
                </div>
            )}
        </Modal>
    )
}

export default Detail
