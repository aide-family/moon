import { FC, ReactNode, useState } from 'react'
import {
    Button,
    Col,
    Form,
    FormInstance,
    Input,
    Row,
    Select,
    Space
} from 'antd'
import PromQLInput, {
    PromValidate,
    formatExpressionFunc
} from '@/components/Prom/PromQLInput.tsx'
import {
    durationOptions,
    getAlarmPages,
    getCategories,
    getEndponts,
    getLevels,
    getRestrain,
    getStrategyGroups,
    maxSuppressUnitOptions,
    strategyEditOptions,
    sverityOptions
} from '../options'
import { DeleteOutlined } from '@ant-design/icons'
import AddLabelModal from './AddLabelModal'
import DataForm from '@/components/Data/DataForm/DataForm'
import FetchSelect from '@/components/Data/FetchSelect'
import { DefaultOptionType } from 'antd/es/select'
import TimeUintInput from './TimeUintInput'
import { Rule } from 'antd/es/form'

export type FormValuesType = {
    alert?: string
    annotations?: {
        title: string
        description: string
        [key: string]: string
    }
    duration?: string
    dataSource?: DefaultOptionType
    groupId?: number
    lables?: { sverity?: string; [key: string]: string | undefined }
    expr?: string
    restrain?: number[]
    alarmPageIds?: number[]
    categoryIds?: number[]
    remark?: string
    levelId?: number
    sendRecover?: boolean
}
export interface StrategyFormProps {
    form: FormInstance
    disabled?: boolean
    groupIdOptions?: DefaultOptionType[]
    alarmPageIdsOptions?: DefaultOptionType[]
    categoryIdsOptions?: DefaultOptionType[]
    endpointOptions?: DefaultOptionType[]
    restrainOptions?: DefaultOptionType[]
    levelOptions?: DefaultOptionType[]
    initialValue?: FormValuesType
}

export type labelsType = {
    label: string | ReactNode
    name: string
}

let timeout: NodeJS.Timeout
export const StrategyForm: FC<StrategyFormProps> = (props) => {
    const {
        disabled,
        form,
        groupIdOptions,
        alarmPageIdsOptions,
        categoryIdsOptions,
        endpointOptions,
        restrainOptions,
        levelOptions
        // initialValue
    } = props

    const handleOnChang = (values: any) => {
        console.log('values', values)
    }

    const [labelFormItemList, setLabelFormItemList] = useState<labelsType[]>([])
    const [annotationFormItemList, setAnnotationFormItemList] = useState<
        labelsType[]
    >([])
    const [addLabelModalOpen, setAddLabelModalOpen] = useState<boolean>(false)
    const [isLabelModalOpen, setIsLabelModalOpen] = useState<boolean>(false)
    const [validatePromQL, setValidatePromQL] = useState<PromValidate>({})

    const dataSource = Form.useWatch<DefaultOptionType>('dataSource', form)

    const fetchValidateExpr = async (value?: string) => {
        if (!value) {
            return
        }

        let msg: PromValidate = {}
        try {
            const resp = await formatExpressionFunc(dataSource?.title, value)
            switch (resp.status) {
                case 'error':
                    msg = {
                        help: `[${resp.errorType}] ${resp.error}`,
                        validateStatus: 'error'
                    }
                    break
                case 'success':
                    msg = {
                        help: `语法校验通过✅`,
                        validateStatus: 'success'
                    }
                    break
            }
        } catch (err: any) {
            msg = {
                help: `${err}`,
                validateStatus: 'error'
            }
        }
        setValidatePromQL(msg)
    }

    const PromQLRule: Rule[] = [
        {
            required: true,
            message: 'PromQL不能为空, 请填写PromQL'
        },
        {
            validator: (
                _: any,
                value: string,
                callback: (error?: string) => void
            ) => {
                if (!value) {
                    return callback()
                }
                if (timeout) {
                    clearTimeout(timeout)
                }

                timeout = setTimeout(() => {
                    fetchValidateExpr(value)
                }, 1000)
                return callback()
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

    const buildPathPrefix = () => {
        // 去除末尾/
        const promPathPrefix = dataSource?.title?.replace(/\/$/, '')
        return promPathPrefix
    }

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
                    onFinish: handleOnChang,
                    layout: 'vertical',
                    disabled: disabled
                    // initialValues: initialValue
                }}
                dataSource={
                    <FetchSelect
                        selectProps={{
                            placeholder: '请选择数据源',
                            labelInValue: true
                        }}
                        width="100%"
                        handleFetch={getEndponts}
                        defaultOptions={endpointOptions}
                    />
                }
                groupId={
                    <FetchSelect
                        width="100%"
                        selectProps={{
                            placeholder: '请选择策略分组'
                        }}
                        handleFetch={getStrategyGroups}
                        defaultOptions={groupIdOptions}
                    />
                }
                levelId={
                    <FetchSelect
                        selectProps={{
                            placeholder: '请选择告警级别'
                        }}
                        width="100%"
                        handleFetch={getLevels}
                        defaultOptions={levelOptions}
                    />
                }
                categoryIds={
                    <FetchSelect
                        selectProps={{
                            placeholder: '请选择告警类型',
                            mode: 'multiple'
                        }}
                        width="100%"
                        handleFetch={getCategories}
                        defaultOptions={categoryIdsOptions}
                    />
                }
                restrain={
                    <FetchSelect
                        selectProps={{
                            placeholder: '请选择抑制对象',
                            mode: 'multiple'
                        }}
                        width="100%"
                        handleFetch={getRestrain}
                        defaultOptions={restrainOptions}
                    />
                }
                alarmPageIds={
                    <FetchSelect
                        selectProps={{
                            placeholder: '请选择告警页面',
                            mode: 'multiple'
                        }}
                        width="100%"
                        handleFetch={getAlarmPages}
                        defaultOptions={alarmPageIdsOptions}
                    />
                }
                duration={
                    <TimeUintInput
                        width="100%"
                        placeholder={['请输入持续时间', '选择单位']}
                        unitOptions={durationOptions}
                    />
                }
                maxSuppress={
                    <TimeUintInput
                        width="100%"
                        placeholder={['请输入最大抑制时间', '选择单位']}
                        unitOptions={maxSuppressUnitOptions}
                    />
                }
                sendInterval={
                    <TimeUintInput
                        width="100%"
                        placeholder={['请输入最大通知间隔时间', '选择单位']}
                        unitOptions={maxSuppressUnitOptions}
                    />
                }
            >
                <Form.Item
                    name="expr"
                    label="PromQL"
                    {...validatePromQL}
                    tooltip={
                        <div>
                            正确的PromQL表达式,
                            用于完成Prometheus报警规则数据匹配
                        </div>
                    }
                    rules={PromQLRule}
                    dependencies={['dataSource']}
                    // initialValue={initialValue?.expr}
                >
                    <PromQLInput
                        disabled={disabled}
                        pathPrefix={buildPathPrefix()}
                        // pathPrefix="http://124.223.104.203:9090/"
                        formatExpression={true}
                    />
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
        </>
    )
}
