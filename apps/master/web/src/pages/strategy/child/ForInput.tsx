import React, { useEffect } from "react";
import { Input, InputNumber, Select } from "@arco-design/web-react";

export type ForInputValue = {
  value?: number;
  unit?: string;
};

export interface ForInputProps {
  onChange?: (value?: string) => void;
  value?: ForInputValue;
  defaultValue?: ForInputValue;
  disabled?: boolean;
}

const ForInput: React.FC<ForInputProps> = (props) => {
  const { onChange, value, defaultValue, disabled } = props;
  const [data, setData] = React.useState<ForInputValue | undefined>(
    value || defaultValue
  );

  const handleOnChange = (value?: number, unit?: string) => {
    setData({ value, unit });
  };

  const unitOptions = [
    { label: "秒(s)", value: "s" },
    { label: "分(m)", value: "m" },
    { label: "时(h)", value: "h" },
    { label: "天(d)", value: "d" },
  ];

  useEffect(() => {
    if (onChange) {
      let val: string | undefined = undefined;
      if (data && data?.value && data?.unit) {
        val = `${data?.value}${data?.unit}`;
      }
      onChange(val);
    }
  }, [data, onChange]);

  return (
    <Input.Group compact style={{ display: "flex" }}>
      <InputNumber
        disabled={disabled}
        size="large"
        placeholder="请输入持续时间"
        style={{ width: "74%" }}
        defaultValue={data?.value}
        onChange={(v) => handleOnChange(v, data?.unit)}
      />
      <Select
        disabled={disabled}
        options={unitOptions}
        size="large"
        style={{ minWidth: "100px", maxWidth: "25%" }}
        defaultValue={data?.unit}
        onChange={(v) => handleOnChange(data?.value, v)}
      />
    </Input.Group>
  );
};

export default ForInput;
