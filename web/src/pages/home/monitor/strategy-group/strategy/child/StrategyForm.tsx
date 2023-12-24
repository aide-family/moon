import React, {FC} from "react";
import {Button, Form} from "antd";
import PromQLInput, {formatExpressionFunc, PromValidate} from "@/components/Prom/PromQLInput.tsx";

export interface StrategyFormProps {}

let timeout: NodeJS.Timeout
export const StrategyForm:FC<StrategyFormProps> = (props) =>{
    console.log(props)
    const [form] = Form.useForm()

    const pathPrefix = 'http://124.223.104.203:30682' //

    const handleOnChang = (values: any) => {
        console.log('values', values)
    }

    const [promValidate, setPromValidate] = React.useState<
        PromValidate | undefined
    >()

    const fetchValidateExpr = async (value?: string) => {
        setPromValidate({
            help: 'Your PromQL is validating',
            validateStatus: 'validating'
        })
        try {
            const resp = await formatExpressionFunc(pathPrefix, value)
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
    return <>
        <Form form={form} onFinish={handleOnChang}>
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
                rules={[
                    {
                        validator: (_, value: string) => {
                            clearTimeout(timeout)
                            if (!value) {
                                setPromValidate({
                                    help: 'PromQL不能为空, 请填写PromQL',
                                    validateStatus: 'error'
                                })
                                return Promise.reject(
                                    'PromQL不能为空, 请填写PromQL'
                                )
                            }
                            timeout = setTimeout(() => {
                                fetchValidateExpr(value)
                            }, 1000)
                            return Promise.resolve()
                        }
                    }
                ]}
            >
                <PromQLInput
                    pathPrefix={pathPrefix}
                    formatExpression={true}
                    promValidate={promValidate}
                />
            </Form.Item>
            <Form.Item>
                <Button htmlType="submit">submit</Button>
            </Form.Item>
        </Form>
    </>
}

