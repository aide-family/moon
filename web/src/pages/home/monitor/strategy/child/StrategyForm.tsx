import { FC, ReactNode, useEffect, useState } from 'react'
import {
    Button,
    Col,
    Form,
    FormInstance,
    Input,
    InputNumber,
    Row,
    Select,
    Space
} from 'antd'
import PromQLInput, {
    formatExpressionFunc,
    PromValidate
} from '@/components/Prom/PromQLInput.tsx'
import {
    alarmPageOptions,
    categoryOptions,
    durationOptions,
    endpoIntOptions,
    restrainOptions,
    strategyGroupOptions,
    sverityOptions
} from '../options'
import { DeleteOutlined } from '@ant-design/icons'
import AddLabelModal from './AddLabelModal'

export type UintType = 's' | 'm' | 'h' | 'd'

export type FormValuesType = {
    alert: string
    annotations: {
        title: string
        description: string
        [key: string]: string
    }
    duration: { value: number; unit: UintType }
    endpoint: string
    groupId: number
    lables: { sverity: string; [key: string]: string }
    expr: string
    restrain: number[]
    alarmPageIds: number[]
    categoryIds: number[]
}
export interface StrategyFormProps {
    form: FormInstance
    disabled?: boolean
    initValues?: FormValuesType
}

export type labelsType = {
    label: string | ReactNode
    name: string
}

let timeout: NodeJS.Timeout
export const StrategyForm: FC<StrategyFormProps> = (props) => {
    const { disabled, form, initValues } = props

    const handleOnChang = (values: any) => {
        console.log('values', values)
    }

    const [promValidate, setPromValidate] = useState<PromValidate | undefined>()
    const [labelFormItemList, setLabelFormItemList] = useState<labelsType[]>([])
    const [annotationFormItemList, setAnnotationFormItemList] = useState<
        labelsType[]
    >([])
    const [addLabelModalOpen, setAddLabelModalOpen] = useState<boolean>(false)
    const [isLabelModalOpen, setIsLabelModalOpen] = useState<boolean>(false)

    const endpoint = Form.useWatch('endpoint', form)

    const fetchValidateExpr = async (value?: string) => {
        setPromValidate({
            help: 'Your PromQL is validating',
            validateStatus: 'validating'
        })
        try {
            const resp = await formatExpressionFunc(endpoint, value)
            setPromValidate({
                help: 'Your PromQL is valid',
                validateStatus: 'success'
            })
            return resp
        } catch (err: any) {
            setPromValidate({
                help: err,
                validateStatus: 'error'
            })
            return err
        }
    }

    const PromQLRule = [
        {
            validator: (_: any, value: string) => {
                clearTimeout(timeout)
                if (!value) {
                    setPromValidate({
                        help: 'PromQL不能为空, 请填写PromQL',
                        validateStatus: 'error'
                    })
                    return Promise.reject('PromQL不能为空, 请填写PromQL')
                }
                timeout = setTimeout(() => {
                    fetchValidateExpr(value)
                }, 1000)
                return Promise.resolve()
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

    useEffect(() => {
        form.setFieldsValue(initValues)
    }, [initValues])

    useEffect(() => {
        if (!endpoint) {
            setPromValidate(undefined)
        }
    }, [endpoint])

    return (
        <>
            <AddLabelModal
                open={addLabelModalOpen}
                onCancel={handleCloseAddLabelModal}
                onOk={handleAddLabel}
            />
            <Form
                form={form}
                onFinish={handleOnChang}
                layout="vertical"
                disabled={disabled}
            >
                <Row gutter={16}>
                    <Col span={12}>
                        <Form.Item
                            name="endpoint"
                            label="数据源"
                            tooltip={
                                <p>
                                    请选择Prometheus数据源, 目前仅支持Prometheus
                                </p>
                            }
                            rules={[
                                {
                                    required: true,
                                    message: '请选择Prometheus数据源'
                                }
                            ]}
                        >
                            <Select
                                disabled={false}
                                allowClear
                                options={endpoIntOptions}
                                placeholder="请选择Prometheus数据源"
                            />
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item
                            name="groupId"
                            label="策略组"
                            tooltip={
                                <>
                                    <p>
                                        把当前规则归类到不同的策略组,
                                        便于业务关联
                                    </p>
                                </>
                            }
                            rules={[
                                {
                                    required: true,
                                    message: '请选择策略组'
                                }
                            ]}
                        >
                            <Select
                                disabled={false}
                                allowClear
                                options={strategyGroupOptions}
                                placeholder="请选择策略组"
                            />
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item
                            name="alert"
                            label="策略名称"
                            tooltip={
                                <>
                                    <p>
                                        请输入策略名称, 策略名称必须唯一, 例如:
                                        'cpu_usage'
                                    </p>
                                </>
                            }
                            rules={[
                                {
                                    required: true,
                                    message: '请输入策略名称'
                                }
                            ]}
                        >
                            <Input placeholder="请输入策略名称" />
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item
                            label="持续时间"
                            tooltip={
                                <>
                                    <p>
                                        持续时间是下面PromQL规则连续匹配,
                                        建议为此规则采集周期的整数倍,
                                        例如采集周期为15s, 持续时间为30s,
                                        则表示连续2个周期匹配
                                    </p>
                                </>
                            }
                            rules={[
                                {
                                    required: true,
                                    message: '请输入持续时间'
                                }
                            ]}
                        >
                            <Space.Compact style={{ width: '100%' }}>
                                <Form.Item
                                    name={['duration', 'value']}
                                    initialValue={3}
                                    noStyle
                                >
                                    <InputNumber
                                        placeholder="请输入持续时间"
                                        style={{ width: '80%' }}
                                    />
                                </Form.Item>
                                <Form.Item
                                    name={['duration', 'unit']}
                                    initialValue="m"
                                    noStyle
                                >
                                    <Select
                                        options={durationOptions}
                                        style={{ width: '20%', minWidth: 80 }}
                                    />
                                </Form.Item>
                            </Space.Compact>
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item
                            name="alarmPageIds"
                            label="报警页面"
                            tooltip={
                                <>
                                    <p>
                                        报警页面: 当该规则触发时,
                                        页面将跳转到报警页面
                                    </p>
                                </>
                            }
                            rules={[
                                {
                                    required: true,
                                    message: '请选择报警页面'
                                }
                            ]}
                        >
                            <Select
                                disabled={false}
                                allowClear
                                options={alarmPageOptions}
                                mode="multiple"
                                placeholder="请选择报警页面"
                            />
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item
                            name="categoryIds"
                            label="策略类型"
                            tooltip={
                                <>
                                    <p>
                                        策略类型: 选择策略类型, 例如:
                                        网络、业务、系统等
                                    </p>
                                </>
                            }
                            rules={[
                                {
                                    required: true,
                                    message: '请选择策略类型'
                                }
                            ]}
                        >
                            <Select
                                disabled={false}
                                allowClear
                                options={categoryOptions}
                                mode="multiple"
                                placeholder="请选择策略类型"
                            />
                        </Form.Item>
                    </Col>
                </Row>

                <Form.Item
                    name="restrain"
                    label="抑制对象"
                    initialValue={[]}
                    tooltip={
                        <div>
                            抑制对象: 当该规则触发时, 此列表对象的告警将会被抑制
                        </div>
                    }
                >
                    <Select
                        mode="multiple"
                        placeholder="请选择抑制对象"
                        options={restrainOptions}
                    />
                </Form.Item>

                <Form.Item
                    name="expr"
                    label="PromQL"
                    {...promValidate}
                    tooltip={
                        <div>
                            正确的PromQL表达式,
                            用于完成Prometheus报警规则数据匹配
                        </div>
                    }
                    rules={PromQLRule}
                    dependencies={['endpoint']}
                >
                    {!!endpoint ? (
                        <PromQLInput
                            disabled={disabled}
                            pathPrefix={endpoint}
                            formatExpression={true}
                            promValidate={promValidate}
                        />
                    ) : (
                        <div>数据源为空, 不予渲染PromQL输入框</div>
                    )}
                </Form.Item>

                <Form.Item
                    tooltip={
                        <div>
                            标签: 标签是Prometheus报警规则的附加信息, 例如:
                            告警等级, 告警实例等, 也可以添加自定义标签
                        </div>
                    }
                    label={
                        <span style={{ color: '#E0E2E6' }}>
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
                                name={['lables', 'sverity']}
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
                                        label={item.label}
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
                            >
                                告警注释
                            </Button>
                            <span style={{ color: '#E0E2E6' }}>(可选)</span>
                        </span>
                    }
                >
                    <Form.Item
                        name={['annotations', 'title']}
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
                    {annotationFormItemList.map((item, index) => {
                        return (
                            <Form.Item
                                name={['annotations', item.name]}
                                label={item.label}
                                rules={[
                                    {
                                        required: true,
                                        message: `请输入${item.label}`
                                    }
                                ]}
                                key={index}
                            >
                                <Input placeholder={`请输入${item.label}`} />
                            </Form.Item>
                        )
                    })}
                </Form.Item>
            </Form>
        </>
    )
}
