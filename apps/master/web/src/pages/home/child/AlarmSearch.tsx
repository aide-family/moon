import React from "react";
import { Form, Input, Select } from "@arco-design/web-react";

import type { AlarmQueryParams } from "@/apis/prom/alarm/alarm";

import styles from "../style/alarm.module.less";

export interface AlarmSearchProps {
  queryParams?: AlarmQueryParams;
  setQueryParamas?: React.Dispatch<React.SetStateAction<AlarmQueryParams>>;
}

const AlarmSearch: React.FC<AlarmSearchProps> = (props) => {
  const { queryParams, setQueryParamas } = props;

  const [form] = Form.useForm();

  const handleSearchFormChange = (
    value: Partial<FormData>,
    values: Partial<FormData>
  ) => {
    // TODO 参数以实际为准, 先占位
    setQueryParamas?.(values as AlarmQueryParams);
  };

  return (
    <>
      <div className={styles.alarmSearch}>
        <Form form={form} layout="inline" onChange={handleSearchFormChange}>
          <Form.Item label="规则ID" field="strategyId">
            <Input
              placeholder="根据规则ID搜索报警..."
              allowClear
              className={styles.input}
            />
          </Form.Item>
          <Form.Item label="报警页面" field="pageId">
            <Select
              options={[]}
              placeholder="选择你要查询的报警页面"
              allowClear
              className={styles.select}
              mode="multiple"
            />
          </Form.Item>
        </Form>
      </div>
    </>
  );
};

export default AlarmSearch;
