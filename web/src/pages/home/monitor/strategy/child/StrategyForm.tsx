import { FC, useEffect, useState } from 'react'
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
import { durationOptions } from '../options'
import { DeleteOutlined } from '@ant-design/icons'

export interface StrategyFormProps {
    form: FormInstance
    disabled?: boolean
}

let timeout: NodeJS.Timeout
export const StrategyForm: FC<StrategyFormProps> = (props) => {
    const { disabled, form } = props

    const handleOnChang = (values: any) => {
        console.log('values', values)
    }

    const [promValidate, setPromValidate] = useState<PromValidate | undefined>()
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

    // TODO 获取数据源
    const options = [
        {
            label: 'Prometheus',
            value: 'http://124.223.104.203:9090'
        },
        {
            label: 'Grafana',
            value: 'http://124.223.104.203:3000'
        }
    ]
    // TODO 获取策略组列表
    const strategyGroupOptions = [
        {
            label: 'Default',
            value: 'default'
        },
        {
            label: '网络',
            value: 'network'
        },
        {
            label: '存储',
            value: 'storage'
        }
    ]

    useEffect(() => {
        if (!endpoint) {
            setPromValidate(undefined)
        }
    }, [endpoint])

    return (
        <>
            <Form
                form={form}
                onFinish={handleOnChang}
                layout="vertical"
                disabled={disabled}
            >
                <Row gutter={16}>
                    <Col span={6}>
                        <Form.Item
                            name="endpoint"
                            label="数据源"
                            tooltip={
                                <>
                                    <p>
                                        请选择Prometheus数据源,
                                        目前仅支持Prometheus
                                    </p>
                                </>
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
                                options={options}
                                placeholder="请选择Prometheus数据源"
                            />
                        </Form.Item>
                    </Col>
                    <Col span={6}>
                        <Form.Item
                            name="strategyGroupId"
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
                    <Col span={6}>
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
                    <Col span={6}>
                        <Form.Item
                            name="duration"
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
                                <InputNumber
                                    placeholder="请输入持续时间"
                                    style={{ width: '80%' }}
                                    defaultValue={3}
                                />
                                <Select
                                    defaultValue="m"
                                    options={durationOptions}
                                    style={{ width: '20%', minWidth: 80 }}
                                />
                            </Space.Compact>
                        </Form.Item>
                    </Col>
                </Row>

                <Form.Item
                    name="restrain"
                    label="抑制对象"
                    tooltip={
                        <div>
                            抑制对象: 当该规则触发时, 此列表对象的告警将会被抑制
                        </div>
                    }
                >
                    <Select mode="multiple" placeholder="请选择抑制对象" />
                </Form.Item>

                <Form.Item
                    name="promQL"
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

                <span style={{ color: '#E0E2E6' }}>
                    <Button type="primary" size="small">
                        添加标签
                    </Button>
                    <span>(可选)</span>
                </span>
                <Form.Item
                    name="sverity"
                    label="等级(sverity)"
                    style={{ marginTop: 12 }}
                    rules={[
                        {
                            required: true,
                            message: '请输入告警等级'
                        }
                    ]}
                >
                    <Space.Compact>
                        <Input placeholder="请输入告警等级" />
                        <Button
                            type="primary"
                            icon={<DeleteOutlined />}
                            danger
                        />
                    </Space.Compact>
                </Form.Item>
                <span style={{ color: '#E0E2E6' }}>
                    <Button type="primary" size="small">
                        告警注释
                    </Button>
                    <span style={{ color: '#E0E2E6' }}>(可选)</span>
                </span>
                <Form.Item
                    name="title"
                    label="告警标题模板"
                    style={{ marginTop: 12 }}
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
                    name="title"
                    label="告警内容模板"
                    style={{ marginTop: 12 }}
                    rules={[
                        {
                            required: true,
                            message: '请输入告警内容模板'
                        }
                    ]}
                >
                    <Input.TextArea placeholder="请输入告警内容模板" />
                </Form.Item>
            </Form>
        </>
    )
}
