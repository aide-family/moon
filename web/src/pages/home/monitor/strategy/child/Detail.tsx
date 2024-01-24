import { FC, useEffect, useState } from 'react'
import { Form, Modal, Spin } from 'antd'
import {
    StrategyCreateRequest,
    StrategyItemType,
    StrategyUpdateRequest
} from '@/apis/home/monitor/strategy/types'
import { FormValuesType, StrategyForm } from './StrategyForm'
import { ActionKey } from '@/apis/data'
import strategyApi from '@/apis/home/monitor/strategy'
import { Duration } from '@/apis/types'

export interface DetailProps {
    open: boolean
    onClose: () => void
    id?: number
    disabled?: boolean
    actionKey?: ActionKey
    refresh?: () => void
}

const buildTimeDuration = (d?: Duration) => {
    return d && d.unit && d.value
        ? {
              unit: d.unit,
              value: +d.value
          }
        : undefined
}

export const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, disabled, actionKey, id, refresh } = props

    const [form] = Form.useForm<FormValuesType>()

    const [detail, setDetail] = useState<StrategyItemType>()
    const [loading, setLoading] = useState<boolean>(false)

    const fetchDetail = () => {
        console.log('fetchDetail', id)
        form?.resetFields()
        setDetail(undefined)
        if (!id) {
            return
        }
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
            duration: strategyFormValues.duration,
            alarmLevelId: strategyFormValues?.alarmLevelId || 0,
            dataSourceId: (strategyFormValues.dataSource?.value as number) || 0,
            labels: strategyFormValues?.lables || {},
            annotations: strategyFormValues?.annotations || {},
            expr: strategyFormValues.expr || '',
            groupId: strategyFormValues?.groupId || 0,
            alert: strategyFormValues.alert || '',
            alarmPageIds: strategyFormValues?.alarmPageIds || [],
            categoryIds: strategyFormValues?.categoryIds || [],
            remark: strategyFormValues.remark || '',
            maxSuppress: strategyFormValues?.maxSuppress,
            sendInterval: strategyFormValues?.sendInterval,
            sendRecover: strategyFormValues?.sendRecover
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
            duration: strategyFormValues.duration,
            alarmLevelId: strategyFormValues.alarmLevelId || 0,
            labels: strategyFormValues.lables || {},
            annotations: strategyFormValues.annotations || {},
            expr: strategyFormValues.expr || '',
            dataSourceId: (strategyFormValues.dataSource?.value as number) || 0,
            maxSuppress: buildTimeDuration(strategyFormValues?.maxSuppress),
            sendInterval: buildTimeDuration(strategyFormValues?.sendInterval),
            sendRecover: strategyFormValues?.sendRecover
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
                console.log('values', values)
                // 单独校验expr字段
                switch (actionKey) {
                    case ActionKey.ADD:
                        handleAddStrategy(values)
                        break
                    case ActionKey.EDIT:
                        handleEditStrategy(values)
                        break
                }
                return values
            })
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        if (!open) {
            return
        }
        fetchDetail()
    }, [open])

    return (
        <Modal
            title={<Title />}
            open={open}
            onCancel={onClose}
            onOk={handleSubmit}
            width="66%"
            centered={true}
            // destroyOnClose={true}
        >
            <Spin spinning={loading}>
                <StrategyForm
                    form={form}
                    disabled={disabled}
                    initialValue={detail}
                />
            </Spin>
        </Modal>
    )
}
