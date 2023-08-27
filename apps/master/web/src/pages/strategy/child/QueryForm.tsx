import React from "react";
import { Form, FormInstance, Input } from "@arco-design/web-react";
import type { PromStrategyItemRequest } from "@/apis/prom/prom";

export interface QueryFormProps {
  form?: FormInstance;
  onChange?: (values: PromStrategyItemRequest) => void;
}
let timer: NodeJS.Timeout;

const QueryForm: React.FC<QueryFormProps> = (props) => {
  const { form, onChange } = props;

  const handleQueryFormOnchan = (_: Partial<any>, values: Partial<any>) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      onChange?.(values);
    }, 500);
  };

  return (
    <>
      <Form layout="inline" form={form} onChange={handleQueryFormOnchan}>
        <Form.Item label="规则名称" field="alert">
          <Input
            placeholder="通过规则名称模糊搜索"
            allowClear
            style={{ width: 600 }}
          />
        </Form.Item>
      </Form>
    </>
  );
};

export default QueryForm;
