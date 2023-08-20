import React, { useEffect, useState } from "react";
import PromQLInput, {
  formatExpressionFunc,
  PromValidate,
} from "@/components/Prom/PromQLInput";
import { Form, RulesProps } from "@arco-design/web-react";

export interface PromQLFormItemProps {
  pathPrefix: string;
  label?: string;
  field?: string;
  placeholder?: string;
  required?: boolean;
  rules?: RulesProps<string>[];
  disabled?: boolean;
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
  const [expr, setExpr] = useState<string | undefined>("");

  const fetchValidateExpr = (value?: string) => {
    formatExpressionFunc(pathPrefix, value)
      .then((resp) => {
        setPromValidate({
          help: "Your PromQL is valid",
          validateStatus: "success",
        });
        return resp;
      })
      .catch((err) => {
        setPromValidate({
          help: err,
          validateStatus: "error",
        });
        return err;
      });
  };

  useEffect(() => {
    fetchValidateExpr(expr);
  }, [pathPrefix]);

  return (
    <Form.Item
      label={label}
      field={field}
      {...promValidate}
      disabled={disabled}
      tooltip={
        <div>正确的PromQL表达式, 用于完成Prometheus报警规则数据匹配</div>
      }
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
              fetchValidateExpr(value);
            }, 1000);
          },
        },
      ]}
    >
      <PromQLInput
        onChange={setExpr}
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
