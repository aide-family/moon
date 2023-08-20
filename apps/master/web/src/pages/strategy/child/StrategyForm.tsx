import React, { useEffect, useState } from "react";
import {
  Button,
  Form,
  FormInstance,
  Grid,
  Input,
  Message,
  Select,
} from "@arco-design/web-react";
import PromQLFormItem from "@/components/Prom/PromQLFormItem";
import MapFormModal, { Map } from "@/pages/strategy/child/MapFormModal";
import { IconDelete } from "@arco-design/web-react/icon";
import type { ForInputValue } from "@/pages/strategy/child/ForInput";
import ForInput from "@/pages/strategy/child/ForInput";
import type { StrategyValues } from "@/pages/strategy/child/StrategyModal";
import type { DatasourceItem } from "@/apis/prom/prom";
import { Datasources } from "@/apis/prom/dict/api";

export interface StrategyModalProps {
  form: FormInstance;
  initialValues?: any;
  onChange?: (values?: StrategyValues) => void;
  disabled?: boolean;
}

const { Row, Col } = Grid;

const pathPrefix = "https://prometheus.bitrecx.com";

const StrategyForm: React.FC<StrategyModalProps> = (props) => {
  const { form, disabled, initialValues, onChange } = props;
  const datasource =
    Form.useWatch("datasource", form) ||
    initialValues?.datasource ||
    pathPrefix;
  const [labels, setLabels] = useState<Map[]>([]);
  const [annotations, setAnnotations] = useState<Map[]>([]);
  const [promDatasource, setPromDatasource] = useState<DatasourceItem[]>([]);

  const getpromDatasource = () => {
    Datasources().then((resp) => {
      const { datasources, response } = resp;
      if (response.code !== "0") {
        return;
      }
      setPromDatasource(datasources || []);
    });
  };

  const setLabelsValue = (value: Map) => {
    setLabels(
      [...labels, value].filter((item, index, arr) => {
        return arr.findIndex((v) => v.name === item.name) === index;
      })
    );
  };

  const setAnnotationsValue = (value: Map) => {
    // 去重
    setAnnotations(
      [...annotations, value].filter((item, index, arr) => {
        return arr.findIndex((v) => v.name === item.name) === index;
      })
    );
  };

  const removeLabel = (index: number) => {
    const field = `labels.${labels?.[index]?.key}`;
    // 删除form中的数据
    form.resetFields([field]);
    const newLabels = [...labels];
    newLabels.splice(index, 1);
    setLabels(newLabels);
    onChange?.(form.getFieldsValue());
  };

  const removeAnnotation = (index: number) => {
    // 删除form中的数据
    form.resetFields([`annotations.${annotations?.[index]?.key}`]);
    const newAnnotations = [...annotations];
    newAnnotations.splice(index, 1);
    setAnnotations(newAnnotations);
    onChange?.(form.getFieldsValue());
  };

  const buildForParam = (forString?: string): ForInputValue => {
    return {
      value: Number(forString?.replace(/[a-zA-Z]/g, "")),
      unit: forString?.replace(/[0-9]/g, ""),
    };
  };

  const buildMap = (
    val: { [key: string]: string },
    setData: React.Dispatch<React.SetStateAction<Map[]>>
  ) => {
    if (!val) return;
    const newMap = Object.keys(val).map(
      (key): { name: string; key: string } => {
        return {
          name: key,
          key: key,
        };
      }
    );

    setData(newMap);
  };

  useEffect(() => {
    if (!initialValues && !form) return;
    buildMap(initialValues?.labels, setLabels);
    buildMap(initialValues?.annotations, setAnnotations);
  }, [initialValues, form]);

  useEffect(() => {
    getpromDatasource();
  }, []);

  return (
    <Form
      form={form}
      layout="vertical"
      autoComplete="off"
      disabled={disabled}
      initialValues={{
        ...initialValues,
        for: buildForParam(initialValues?.["for"]),
      }}
      onChange={(_, values: Partial<StrategyValues>) =>
        onChange && onChange(values)
      }
    >
      <Form.Item
        label="数据源"
        field="datasource"
        disabled={false}
        tooltip={<div>选择合适的数据源, 用于校验该PromQL语句正确性</div>}
        rules={[
          {
            required: true,
            message: "数据源不能为空, 请填写数据源",
          },
        ]}
      >
        <Select
          placeholder="请输入数据源"
          disabled={false}
          options={[
            ...promDatasource.map((item) => ({
              label: `${item.name}(${item.url})`,
              value: item.url,
            })),
          ]}
        />
      </Form.Item>
      <Row
        gutter={16}
        style={{
          margin: 0,
          padding: 0,
        }}
      >
        <Col span={16}>
          <Form.Item
            label="策略名称"
            field="alert"
            rules={[
              {
                required: true,
                message: "策略名称不能为空, 请填写策略名称",
              },
            ]}
          >
            <Input size="large" placeholder="请输入策略名称" />
          </Form.Item>
        </Col>
        <Col span={8}>
          <Form.Item
            label="持续时间"
            field="for"
            rules={[
              {
                required: true,
                message: "持续时间不能为空, 请填写持续时间",
              },
            ]}
          >
            <ForInput disabled={disabled} />
          </Form.Item>
        </Col>
      </Row>
      <PromQLFormItem
        disabled={disabled}
        pathPrefix={datasource}
        field="expr"
        label="PromQL"
        placeholder="请输入PromQL"
        rules={[
          {
            required: true,
            message: "PromQL不能为空, 请编写PromQL语句",
          },
        ]}
      />
      <Form.Item
        disabled={disabled}
        label={
          <span>
            <MapFormModal
              disabled={disabled}
              onFinished={(map) => setLabelsValue(map)}
              title="添加标签"
            />
            <span style={{ color: "var(--color-neutral-4)" }}> (可选)</span>
          </span>
        }
      >
        <Row gutter={16}>
          {labels.map((item, index) => {
            return (
              <Col span={12} key={index}>
                <Form.Item
                  layout="vertical"
                  key={index}
                  field={`labels.${item.key}`}
                  label={`${item.name}(${item.key})`}
                  rules={[
                    {
                      required: true,
                      message: `${item.name}不能为空, 请填写${item.name}`,
                    },
                  ]}
                >
                  <Input
                    size="large"
                    placeholder={`请输入${item.name}(${item.key})`}
                    suffix={
                      <Button
                        disabled={disabled}
                        type="primary"
                        status="danger"
                        icon={<IconDelete />}
                        style={{ position: "absolute", right: 0 }}
                        onClick={() => removeLabel(index)}
                        size="large"
                      />
                    }
                  />
                </Form.Item>
              </Col>
            );
          })}
        </Row>
      </Form.Item>
      <Form.Item
        label={
          <span>
            <MapFormModal
              disabled={disabled}
              onFinished={(map) => setAnnotationsValue(map)}
              title="添加注释"
            />
            <span style={{ color: "var(--color-neutral-4)" }}> (可选)</span>
          </span>
        }
      >
        <Row gutter={16}>
          {annotations.map((item, index) => {
            return (
              <Col span={12} key={index}>
                <Form.Item
                  layout="vertical"
                  key={index}
                  field={`annotations.${item.key}`}
                  label={`${item.name}(${item.key})`}
                  rules={[
                    {
                      required: true,
                      message: `${item.name}不能为空, 请填写${item.name}`,
                    },
                  ]}
                >
                  <Input
                    size="large"
                    placeholder={`请输入${item.name}(${item.key})`}
                    suffix={
                      <Button
                        disabled={disabled}
                        type="primary"
                        status="danger"
                        icon={<IconDelete />}
                        style={{ position: "absolute", right: 0 }}
                        onClick={() => removeAnnotation(index)}
                        size="large"
                      />
                    }
                  />
                </Form.Item>
              </Col>
            );
          })}
        </Row>
      </Form.Item>
    </Form>
  );
};

export default StrategyForm;
