import React from "react";
import { DatePicker, Form, InputNumber } from "@arco-design/web-react";
import type { ShortcutType } from "@arco-design/web-react/es/DatePicker/interface";
import dayjs from "dayjs";

export type DateDataType = "date" | "range";

export type SearchFormProps<T = number | [number, number]> = {
  onSearch?: (type: DateDataType, value: T, step?: number) => void;
  type?: DateDataType;
};

const shortcutsRangePicker: ShortcutType[] = [
  {
    text: "近一小时",
    value: () => [dayjs().subtract(1, "hour"), dayjs()],
  },
  {
    text: "近3小时",
    value: () => [dayjs().subtract(3, "hour"), dayjs()],
  },
  {
    text: "近6小时",
    value: () => [dayjs().subtract(6, "hour"), dayjs()],
  },
  {
    text: "近12小时",
    value: () => [dayjs().subtract(12, "hour"), dayjs()],
  },
  {
    text: "最近一天",
    value: () => [dayjs().subtract(1, "day"), dayjs()],
  },
  {
    text: "今天",
    value: () => [dayjs().startOf("day"), dayjs().endOf("day")],
  },
  {
    text: "昨天",
    value: () => [
      dayjs().subtract(1, "day").startOf("day"),
      dayjs().subtract(1, "day").endOf("day"),
    ],
  },
  {
    text: "最近三天",
    value: () => [dayjs().subtract(3, "day"), dayjs()],
  },
  {
    text: "最近一周",
    value: () => [dayjs().subtract(1, "week"), dayjs()],
  },
];

const shortcutsDatePicker: ShortcutType[] = [
  {
    text: "此刻",
    value: () => dayjs(),
  },
  {
    text: "一小时前",
    value: () => dayjs().subtract(1, "hour"),
  },
  {
    text: "三小时前",
    value: () => dayjs().subtract(3, "hour"),
  },
  {
    text: "六小时前",
    value: () => dayjs().subtract(6, "hour"),
  },
  {
    text: "十二小时前",
    value: () => dayjs().subtract(12, "hour"),
  },
  {
    text: "一天前",
    value: () => dayjs().subtract(1, "day"),
  },
  {
    text: "三天前",
    value: () => dayjs().subtract(3, "day"),
  },
  {
    text: "一周前",
    value: () => dayjs().subtract(1, "week"),
  },
];

const SearchForm: React.FC<SearchFormProps> = (props) => {
  const { onSearch, type } = props;
  const [form] = Form.useForm();

  const handleOnChang = (value: Partial<any>, values: Partial<any>) => {
    console.log(value, values);
    switch (type) {
      case "range":
        onSearch?.(
          type,
          [
            dayjs(values.date_range[0]).unix(),
            dayjs(values.date_range[1]).unix(),
          ],
          value.step
        );
        break;
      case "date":
        onSearch?.(type, dayjs(values.date).unix());
        break;
      default:
        break;
    }
  };

  return (
    <Form layout="inline" form={form} onChange={handleOnChang}>
      {props.type === "range" ? (
        <>
          <Form.Item
            field="date_range"
            label="时间范围"
            initialValue={[dayjs().subtract(1, "hour"), dayjs()]}
          >
            <DatePicker.RangePicker
              showTime
              style={{ width: 380 }}
              shortcutsPlacementLeft
              shortcuts={shortcutsRangePicker}
            />
          </Form.Item>
          <Form.Item field="step" initialValue={14} label="步长">
            <InputNumber min={1} max={60} step={1} />
          </Form.Item>
        </>
      ) : (
        <Form.Item field="date" label="时间" initialValue={dayjs()}>
          <DatePicker
            showTime
            style={{ width: 380 }}
            shortcutsPlacementLeft
            shortcuts={shortcutsDatePicker}
          />
        </Form.Item>
      )}
    </Form>
  );
};

export default SearchForm;
