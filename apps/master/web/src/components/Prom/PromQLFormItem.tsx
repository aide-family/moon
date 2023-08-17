import React from "react";
import PromQLInput, {formatExpressionFunc, PromValidate,} from "@/components/Prom/PromQLInput";
import {Form, RulesProps} from "@arco-design/web-react";

export interface PromQLFormItemProps {
    pathPrefix: string;
    label?: string;
    field?: string;
    placeholder?: string;
    required?: boolean;
    rules?: RulesProps<string>[];
    disabled?: boolean
}

const PromQLFormItem: React.FC<PromQLFormItemProps> = (props) => {
    const {
        pathPrefix,
        label = "PromQL",
        field = "prom_ql",
        placeholder,
        rules = [],
        required = false,
        disabled,
    } = props;
    let timeout: NodeJS.Timeout;
    const [promValidate, setPromValidate] = React.useState<
        PromValidate | undefined
    >();

    return (
        <Form.Item
            label={label}
            field={field}
            {...promValidate}
            disabled={disabled}
            rules={[
                ...rules,
                {
                    validator: (value) => {
                        clearTimeout(timeout);
                        if (required && !value) {
                            setPromValidate({
                                help: "PromQL不能为空, 请填写PromQL",
                                validateStatus: "error",
                            });
                            return;
                        }
                        timeout = setTimeout(() => {
                            formatExpressionFunc(pathPrefix, value)
                                .then(() =>
                                    setPromValidate({
                                        help: "Your PromQL is valid",
                                        validateStatus: "success",
                                    })
                                )
                                .catch((err) =>
                                    setPromValidate({
                                        help: err,
                                        validateStatus: "error",
                                    })
                                );
                        }, 1000);
                    },
                },
            ]}
        >
            <PromQLInput
                disabled={disabled}
                pathPrefix={pathPrefix}
                formatExpression={true}
                setPromValidate={setPromValidate}
                btnDisabled={promValidate?.validateStatus !== "success"}
                placeholderString={placeholder}
            />
        </Form.Item>
    );
};

export default PromQLFormItem;
