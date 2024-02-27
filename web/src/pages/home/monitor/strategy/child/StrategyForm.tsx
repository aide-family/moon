import { FC, ReactNode, useEffect, useRef, useState } from 'react'
import {
    Button,
    Col,
    Form,
    FormInstance,
    Image,
    Input,
    Modal,
    Row,
    Select,
    Space,
    Tour,
    message
} from 'antd'
import PromQLInput, {
    PromValidate,
    formatExpressionFunc
} from '@/components/Prom/PromQLInput.tsx'
import { strategyEditOptions, sverityOptions, tourSteps } from '../options'
import { DeleteOutlined } from '@ant-design/icons'
import AddLabelModal from './AddLabelModal'
import DataForm from '@/components/Data/DataForm/DataForm'
import { DefaultOptionType } from 'antd/es/select'
import { Rule } from 'antd/es/form'
import { StrategyItemType } from '@/apis/home/monitor/strategy/types'
import { Duration } from '@/apis/types'

export type FormValuesType = {
    alert?: string
    annotations?: {
        summary: string
        description: string
        [key: string]: string
    }
    duration?: Duration
    dataSource?: DefaultOptionType
    groupId?: number
    labels?: { sverity?: string; [key: string]: string | undefined }
    expr?: string
    restrain?: number[]
    alarmPageIds?: number[]
    categoryIds?: number[]
    remark?: string
    alarmLevelId?: number
    dataSourceId: number
    // 最大抑制时常
    maxSuppress?: Duration
    // 告警通知间隔
    sendInterval?: Duration
    // 是否发送告警通知
    sendRecover?: boolean
}
export interface StrategyFormProps {
    form: FormInstance
    disabled?: boolean
    initialValue?: StrategyItemType
    openTour?: boolean
    handleCloseTour?: () => void
}

export type labelsType = {
    label: string | ReactNode
    name: string
}

let timeout: NodeJS.Timeout
export const StrategyForm: FC<StrategyFormProps> = (props) => {
    const promQLRef = useRef(null)
    const promQLButtonRef = useRef(null)
    const labelsRef = useRef(null)
    const annotationsRef = useRef(null)
    const annotationsTitleRef = useRef(null)
    const annotationsDescriptionRef = useRef(null)
    const { disabled, form, initialValue, openTour, handleCloseTour } = props

    const [labelFormItemList, setLabelFormItemList] = useState<labelsType[]>([])
    const [annotationFormItemList, setAnnotationFormItemList] = useState<
        labelsType[]
    >([])
    const [addLabelModalOpen, setAddLabelModalOpen] = useState<boolean>(false)
    const [isLabelModalOpen, setIsLabelModalOpen] = useState<boolean>(false)
    const [validatePromQL, setValidatePromQL] = useState<PromValidate>({})
    const [openCloseTourModal, setOpenCloseTourModal] = useState<boolean>(false)

    const dataSource = Form.useWatch<DefaultOptionType>('dataSource', form)

    const buildInitvalue = () => {
        const value = initialValue
        if (!value) {
            form?.resetFields()
            return
        }
        const init: FormValuesType = {
            ...value,
            labels: {
                ...value?.labels,
                sverity: value?.alarmLevelId
                    ? value.alarmLevelId + ''
                    : undefined
            },
            annotations: {
                ...value?.annotations,
                summary: value?.annotations?.['summary'],
                description: value?.annotations?.['description']
            },
            dataSource: {
                value: value.dataSource?.value,
                label: value.dataSource?.label,
                title: value.dataSource?.endpoint
            },
            restrain: [],
            alert: value?.alert,
            duration: value?.duration,
            alarmLevelId: value?.alarmLevelId,
            alarmPageIds: value?.alarmPageIds,
            expr: value?.expr,
            groupId: value?.groupId,
            categoryIds: value?.categoryIds,
            sendRecover: false
        }
        form?.setFieldsValue(init)
    }

    const handleOnFinishTour = () => {
        handleCloseTour?.()
        setOpenCloseTourModal(false)
        message.success('恭喜你, 已经成功学会了配置prometheus告警规则')
    }

    const handleOpenCloseTourModal = () => {
        setOpenCloseTourModal(true)
    }

    const onHandleCloseTour = () => {
        handleOpenCloseTourModal()
    }

    const handleCloseTourModal = () => {
        setOpenCloseTourModal(false)
    }

    const handleCloseTourModalOnOk = () => {
        setOpenCloseTourModal(false)
        handleCloseTour?.()
        message.success('留花不发待君归')
    }

    const fetchValidateExpr = (value?: string) => {
        if (timeout) {
            clearTimeout(timeout)
        }
        if (!dataSource || !dataSource.title) {
            setValidatePromQL({
                help: '请先选择数据源',
                validateStatus: 'error'
            })
            return Promise.resolve()
        }

        if (!value) {
            setValidatePromQL({
                help: '请填写PromQL',
                validateStatus: 'error'
            })
            return Promise.resolve()
        }

        timeout = setTimeout(() => {
            formatExpressionFunc(dataSource?.title, value)
                .then((resp) => {
                    if (resp.status === 'error') {
                        const msg = `[${resp.errorType}] ${resp.error}`
                        setValidatePromQL({
                            help: `[${resp.errorType}] ${resp.error}`,
                            validateStatus: 'error'
                        })
                        return Promise.reject(new Error(msg))
                    }
                    if (resp.status === 'success') {
                        setValidatePromQL({
                            help: '语法校验通过✅',
                            validateStatus: 'success'
                        })
                    }
                    return resp
                })
                .catch((err: any) => {
                    setValidatePromQL({
                        help: `${err}`,
                        validateStatus: 'error'
                    })
                    return err
                })
        }, 200)

        return Promise.resolve()
    }

    const PromQLRule: Rule[] = [
        {
            required: true,
            message: 'PromQL不能为空, 请填写PromQL'
        },
        {
            validator: (_: Rule, value: string) => {
                return fetchValidateExpr(value)
            }
        }
    ]

    const handleAddLabel = (data: labelsType) => {
        if (data.label && data.name) {
            if (isLabelModalOpen) {
                setLabelFormItemList([...labelFormItemList, data])
            } else {
                setAnnotationFormItemList([...annotationFormItemList, data])
            }
            setAddLabelModalOpen(false)
        }
    }

    const handleCloseAddLabelModal = () => {
        setAddLabelModalOpen(false)
    }

    const openAddLabelModal = () => {
        setAddLabelModalOpen(true)
        setIsLabelModalOpen(true)
    }

    const openAddAnnotationModal = () => {
        setAddLabelModalOpen(true)
        setIsLabelModalOpen(false)
    }

    const handleDeleteLabelFormItemListByIndex = (index: number) => {
        setLabelFormItemList(labelFormItemList.filter((_, i) => i !== index))
    }

    const handleDeleteAnnotationFormItemListByIndex = (index: number) => {
        setAnnotationFormItemList(
            annotationFormItemList.filter((_, i) => i !== index)
        )
    }

    useEffect(() => {
        buildInitvalue()
    }, [initialValue])

    return (
        <>
            <AddLabelModal
                open={addLabelModalOpen}
                onCancel={handleCloseAddLabelModal}
                onOk={handleAddLabel}
            />
            <DataForm
                form={form}
                items={strategyEditOptions}
                formProps={{
                    layout: 'vertical',
                    disabled: disabled
                }}
            >
                <div ref={promQLRef}>
                    <Form.Item
                        name="expr"
                        label="PromQL"
                        {...validatePromQL}
                        required
                        tooltip={
                            <div>
                                正确的PromQL表达式,
                                用于完成Prometheus报警规则数据匹配
                            </div>
                        }
                        rules={PromQLRule}
                        dependencies={['dataSource']}
                    >
                        <PromQLInput
                            disabled={disabled}
                            pathPrefix={dataSource?.title}
                            formatExpression={true}
                            buttonRef={promQLButtonRef}
                        />
                    </Form.Item>
                </div>

                <Form.Item
                    tooltip={
                        <div>
                            标签: 标签是Prometheus报警规则的附加信息, 例如:
                            告警等级, 告警实例等, 也可以添加自定义标签
                        </div>
                    }
                    label={
                        <span style={{ color: '#E0E2E6' }} ref={labelsRef}>
                            <Button
                                type="primary"
                                size="small"
                                onClick={openAddLabelModal}
                            >
                                告警标签
                            </Button>
                            <span>(可选)</span>
                        </span>
                    }
                >
                    <Row gutter={16}>
                        <Col span={6}>
                            <Form.Item
                                name={['labels', 'sverity']}
                                label="等级(sverity)"
                                rules={[
                                    {
                                        required: true,
                                        message: '请输入告警等级'
                                    }
                                ]}
                            >
                                <Select
                                    placeholder="请选择告警等级"
                                    options={sverityOptions}
                                />
                            </Form.Item>
                        </Col>
                        {labelFormItemList.map((item, index) => {
                            return (
                                <Col span={6} key={index}>
                                    <Form.Item
                                        name={['lables', item.name]}
                                        label={`${item.label}(${item.name})`}
                                        rules={[
                                            {
                                                required: true,
                                                message: `请输入${item.label}`
                                            }
                                        ]}
                                    >
                                        <Space.Compact>
                                            <Input
                                                placeholder={`请输入${item.label}`}
                                            />
                                            <Button
                                                type="primary"
                                                danger
                                                icon={<DeleteOutlined />}
                                                onClick={() =>
                                                    handleDeleteLabelFormItemListByIndex(
                                                        index
                                                    )
                                                }
                                            />
                                        </Space.Compact>
                                    </Form.Item>
                                </Col>
                            )
                        })}
                    </Row>
                </Form.Item>

                <Form.Item
                    tooltip={
                        <div>
                            告警注释: 告警注释是Prometheus报警规则的附加信息,
                            例如: 告警标题, 告警描述等, 也可以添加自定义注释
                        </div>
                    }
                    label={
                        <span style={{ color: '#E0E2E6' }}>
                            <Button
                                type="primary"
                                size="small"
                                onClick={openAddAnnotationModal}
                                ref={annotationsRef}
                            >
                                告警注释
                            </Button>
                            <span style={{ color: '#E0E2E6' }}>(可选)</span>
                        </span>
                    }
                >
                    <div ref={annotationsTitleRef}>
                        <Form.Item
                            name={['annotations', 'summary']}
                            label="告警标题模板"
                            rules={[
                                {
                                    required: true,
                                    message: '请输入告警标题模板'
                                }
                            ]}
                        >
                            <Input.TextArea placeholder="请输入告警标题模板" />
                        </Form.Item>
                    </div>
                    <div ref={annotationsDescriptionRef}>
                        <Form.Item
                            name={['annotations', 'description']}
                            label="告警内容模板"
                            rules={[
                                {
                                    required: true,
                                    message: '请输入告警内容模板'
                                }
                            ]}
                        >
                            <Input.TextArea placeholder="请输入告警内容模板" />
                        </Form.Item>
                    </div>

                    {annotationFormItemList.map((item, index) => {
                        return (
                            <Form.Item
                                name={['annotations', item.name]}
                                label={`${item.label}(${item.name})`}
                                rules={[
                                    {
                                        required: true,
                                        message: `请输入${item.label}`
                                    }
                                ]}
                                key={index}
                            >
                                <Space.Compact style={{ width: '100%' }}>
                                    <Input.TextArea
                                        placeholder={`请输入${item.label}`}
                                    />
                                    <Button
                                        type="primary"
                                        danger
                                        icon={<DeleteOutlined />}
                                        onClick={() =>
                                            handleDeleteAnnotationFormItemListByIndex(
                                                index
                                            )
                                        }
                                    />
                                </Space.Compact>
                            </Form.Item>
                        )
                    })}
                </Form.Item>
            </DataForm>
            <Modal
                open={openCloseTourModal}
                onCancel={handleCloseTourModal}
                onOk={handleCloseTourModalOnOk}
                centered
                zIndex={9999999}
            >
                <div
                    style={{
                        width: '100%',
                        height: '100%',
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'center'
                    }}
                >
                    <Image
                        preview={false}
                        src="https://t12.baidu.com/it/u=4123056481,180954535&fm=30&app=106&f=JPEG?w=640&h=360&s=B260C5A41C12A9D45CCC59390300C050"
                    />
                </div>
            </Modal>
            <Tour
                open={openTour}
                onClose={onHandleCloseTour}
                steps={tourSteps({
                    promQLRef,
                    promQLButtonRef,
                    labelsRef,
                    annotationsRef,
                    annotationsTitleRef,
                    annotationsDescriptionRef
                })}
                onFinish={handleOnFinishTour}
            />
        </>
    )
}
