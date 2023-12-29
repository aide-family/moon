import { FC, useEffect, useState } from 'react'
import { Form, Modal, Spin } from 'antd'
import {
    StrategyCreateRequest,
    StrategyItemType,
    StrategyUpdateRequest
} from '@/apis/home/monitor/strategy/types'
import { FormValuesType, StrategyForm, UintType } from './StrategyForm'
import { ActionKey } from '@/apis/data'
import strategyApi from '@/apis/home/monitor/strategy'

export interface DetailProps {
    open: boolean
    onClose: () => void
    id?: number
    disabled?: boolean
    actionKey?: ActionKey
    refresh?: () => void
}

export const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, disabled, actionKey, id, refresh } = props

    const [form] = Form.useForm<FormValuesType>()

    const [detail, setDetail] = useState<StrategyItemType>()
    const [loading, setLoading] = useState<boolean>(false)

    const fetchDetail = () => {
        if (!id) return
        setLoading(true)
        strategyApi
            .getStrategyDetail(id)
            .then((detail) => {
                setDetail(detail)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const Title = () => {
        switch (actionKey) {
            case ActionKey.ADD:
                return '添加策略'
            case ActionKey.EDIT:
                return '编辑策略'
            default:
                return '策略详情'
        }
    }

    const handleAddStrategy = (strategyFormValues: FormValuesType) => {
        const strategyInfo: StrategyCreateRequest = {
            ...strategyFormValues,
            duration:
                strategyFormValues.duration.value +
                strategyFormValues.duration.unit,
            alarmLevelId: +strategyFormValues.lables.sverity,
            dataSourceId: 1,
            labels: strategyFormValues.lables,
            annotations: strategyFormValues.annotations
        }
        strategyApi.addStrategy(strategyInfo).then(() => {
            onClose()
            refresh?.()
        })
    }

    const handleEditStrategy = (strategyFormValues: FormValuesType) => {
        if (!detail) return
        const strategyInfo: StrategyUpdateRequest = {
            ...detail,
            ...strategyFormValues,
            duration:
                strategyFormValues.duration.value +
                strategyFormValues.duration.unit,
            alarmLevelId: +strategyFormValues.lables.sverity,
            labels: strategyFormValues.lables,
            annotations: strategyFormValues.annotations
        }
        strategyApi.updateStrategy(strategyInfo).then(() => {
            onClose()
            refresh?.()
        })
    }

    const handleSubmit = () => {
        console.log(actionKey)
        setLoading(true)
        form.validateFields()
            .then((values) => {
                switch (actionKey) {
                    case ActionKey.ADD:
                        handleAddStrategy(values)
                        break
                    case ActionKey.EDIT:
                        console.log('edit: ', 123)
                        handleEditStrategy(values)
                        break
                }
                return values
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const buildInitvalue = (): FormValuesType => {
        if (!detail) return {} as FormValuesType
        // duration : 1s, 1m, 1h, 1d => [1, s], [1, m], [1, h], [1, d]
        // 截取字符串最后一个字符作为单位, 前面的字符作为数值
        const value = parseInt(detail.duration.slice(0, -1))
        const unit = detail.duration.slice(-1) as UintType
        return {
            ...detail,
            lables: { ...detail?.labels, sverity: `${detail?.alarmLevelId}` },
            annotations: {
                ...detail?.annotations,
                title: detail?.annotations['title'],
                description: detail?.annotations['description']
            },
            endpoint: 'http://124.223.104.203:9090',
            restrain: [1],
            alert: detail?.alert,
            duration: {
                value: value,
                unit: unit
            }
        }
    }

    useEffect(() => {
        form.resetFields()
        setDetail(undefined)
        fetchDetail()
    }, [open])

    return (
        <Modal
            title={<Title />}
            open={open}
            onCancel={onClose}
            onOk={handleSubmit}
            width="66%"
            destroyOnClose={true}
        >
            <Spin spinning={loading}>
                <StrategyForm
                    form={form}
                    disabled={disabled}
                    initValues={buildInitvalue()}
                />
            </Spin>
        </Modal>
    )
}
