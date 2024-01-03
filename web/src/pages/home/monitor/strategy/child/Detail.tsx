import { FC, useEffect, useState } from 'react'
import { Form, Modal, Spin, Tag } from 'antd'
import {
    StrategyCreateRequest,
    StrategyItemType,
    StrategyUpdateRequest
} from '@/apis/home/monitor/strategy/types'
import { FormValuesType, StrategyForm } from './StrategyForm'
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
                form.setFieldsValue(buildInitvalue(detail))
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
            duration: strategyFormValues.duration || '',
            alarmLevelId: strategyFormValues?.lables?.sverity
                ? +strategyFormValues.lables.sverity
                : 0,
            dataSourceId: 1,
            labels: strategyFormValues?.lables || {},
            annotations: strategyFormValues?.annotations || {},
            expr: strategyFormValues.expr || '',
            groupId: strategyFormValues?.groupId || 0,
            alert: strategyFormValues.alert || '',
            alarmPageIds: strategyFormValues?.alarmPageIds || [],
            categoryIds: strategyFormValues?.categoryIds || [],
            remark: strategyFormValues.remark || ''
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
            duration: strategyFormValues.duration || '',
            alarmLevelId: strategyFormValues?.lables?.sverity
                ? +strategyFormValues?.lables?.sverity
                : 0,
            labels: strategyFormValues.lables || {},
            annotations: strategyFormValues.annotations || {},
            expr: strategyFormValues.expr || ''
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

    const buildInitvalue = (value: StrategyItemType): FormValuesType => ({
        ...value,
        lables: {
            ...value?.labels,
            sverity: value?.alarmLevelId ? value.alarmLevelId + '' : undefined
        },
        annotations: {
            ...value?.annotations,
            title: value?.annotations['title'],
            description: value?.annotations['description']
        },
        endpoint: {
            value: 0,
            label: 'localhost',
            title: 'http://localhost:9090'
        },
        restrain: [],
        alert: value?.alert,
        duration: value?.duration,
        levelId: value?.alarmLevelId,
        alarmPageIds: value?.alarmPageIds,
        expr: value?.expr,
        groupId: value?.groupId,
        categoryIds: value?.categoryIds
    })

    const buildAlamrPageIdsOptions = () => {
        if (!detail?.alarmLevelInfo) return []
        return detail.alarmPageInfo?.map((item) => {
            const { color, value, label } = item
            return {
                value: value,
                label: <Tag color={color}>{label}</Tag>
            }
        })
    }

    const buildGroupIdOptions = () => {
        if (!detail?.groupId) return []
        return [
            {
                value: detail?.groupId,
                label: (
                    <Tag color="blue">{detail?.groupInfo?.label || '默认'}</Tag>
                )
            }
        ]
    }

    const categoryIdsOptions = () => {
        if (!detail?.categoryIds) return []
        return detail?.categoryInfo?.map((item) => {
            const { color, value, label } = item
            return {
                value: value,
                label: <Tag color={color}>{label}</Tag>
            }
        })
    }
    const buildEndpointOptions = () => {
        if (!detail?.endpoint) {
            return [
                {
                    value: 0,
                    label: <Tag color="blue">localhost</Tag>,
                    title: 'http://localhost:9090'
                }
            ]
        }
        const { value, label, endpoint } = detail?.endpoint
        return [
            {
                value: value,
                label: <Tag color="blue">{label || '未知'}</Tag>,
                title: endpoint
            }
        ]
    }

    const buildLevelOptions = () => {
        if (!detail?.alarmLevelInfo) return []
        const { value, color, label } = detail?.alarmLevelInfo
        return [
            {
                value: value,
                label: <Tag color={color}>{label}</Tag>
            }
        ]
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
                    groupIdOptions={buildGroupIdOptions()}
                    alarmPageIdsOptions={buildAlamrPageIdsOptions()}
                    categoryIdsOptions={categoryIdsOptions()}
                    endpointOptions={buildEndpointOptions()}
                    levelOptions={buildLevelOptions()}
                />
            </Spin>
        </Modal>
    )
}
