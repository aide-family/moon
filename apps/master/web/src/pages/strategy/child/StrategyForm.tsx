import React, { useEffect, useState } from "react";
import {
  Button,
  Form,
  FormInstance,
  Grid,
  Input,
  Select,
  Tag,
} from "@arco-design/web-react";
import PromQLFormItem from "@/components/Prom/PromQLFormItem";
import MapFormModal, { Map } from "@/pages/strategy/child/MapFormModal";
import { IconDelete } from "@arco-design/web-react/icon";
import type { ForInputValue } from "@/pages/strategy/child/ForInput";
import ForInput from "@/pages/strategy/child/ForInput";
import type { StrategyValues } from "@/pages/strategy/child/StrategyModal";
import type { AlarmPage, DatasourceItem, PromDict } from "@/apis/prom/prom";
import { Category, CategoryMap } from "@/apis/prom/prom";
import { Datasources, DictList } from "@/apis/prom/dict/api";
import { defaultListDictRequest } from "@/apis/prom/dict/dict";
import GroupSelect from "@/pages/strategy/child/GroupSelect";
import { AlarmPageSimpleList } from "@/apis/prom/alarm/api";

export interface StrategyModalProps {
  form: FormInstance;
  initialValues?: StrategyValues;
  onChange?: (values?: StrategyValues) => void;
  disabled?: boolean;
}

const { Row, Col } = Grid;

const StrategyForm: React.FC<StrategyModalProps> = (props) => {
  const { form, disabled, initialValues, onChange } = props;
  const datasource =
    Form.useWatch("datasource", form) || initialValues?.datasource;
  const [labels, setLabels] = useState<Map[]>([]);
  const [annotations, setAnnotations] = useState<Map[]>([]);
  const [promDatasource, setPromDatasource] = useState<DatasourceItem[]>([]);
  const [promAlertLeves, setPromAlertLeves] = useState<PromDict[]>([]);
  const [promCategories, setPromCategories] = useState<PromDict[]>([]);
  const [promAlarmPages, setPromAlarmPages] = useState<AlarmPage[]>([]);

  const getPromDatasource = () => {
    Datasources().then((resp) => {
      const { datasources, response } = resp;
      if (response.code !== "0") {
        return;
      }
      setPromDatasource(datasources || []);
    });
  };

  const getPromAlarmPages = () => {
    AlarmPageSimpleList({
      page: {
        current: 1,
        size: 10,
      },
    }).then((resp) => {
      const { alarmPages, response } = resp;
      if (response.code !== "0") {
        return;
      }
      setPromAlarmPages(alarmPages || []);
    });
  };

  const getDicts = (category: Category) => {
    return DictList({
      ...defaultListDictRequest,
      dict: {
        category: CategoryMap[category],
      },
    });
  };

  const getPromAlertLeves = () => {
    getDicts(Category.CATEGORY_ALERT_LEVEL).then((resp) => {
      const { dicts, response } = resp;
      if (response.code !== "0") return;
      setPromAlertLeves(dicts || []);
    });
  };

  const getPromCategories = () => {
    getDicts(Category.CATEGORY_STRATEGY).then((resp) => {
      const { dicts, response } = resp;
      if (response.code !== "0") return;
      setPromCategories(dicts || []);
    });
  };

  const setLabelsValue = (value: Map) => {
    setLabels(
      [...labels, value].filter((item, index, arr) => {
        return arr.findIndex((v) => v.key === item.key) === index;
      })
    );
  };

  const setAnnotationsValue = (value: Map) => {
    // 去重
    setAnnotations(
      [...annotations, value].filter((item, index, arr) => {
        return arr.findIndex((v) => v.key === item.key) === index;
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
        return { name: key, key: key };
      }
    );

    setData(newMap);
  };

  useEffect(() => {
    if (!initialValues || !form) return;
    if (initialValues.labels) {
      buildMap(initialValues?.labels, setLabels);
    }
    if (initialValues.annotations) {
      buildMap(initialValues?.annotations, setAnnotations);
    }
  }, [initialValues, form]);

  useEffect(() => {
    getPromDatasource();
    getPromAlertLeves();
    getPromCategories();
    getPromAlarmPages();
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
      <Row gutter={16} style={{ margin: 0, padding: 0 }}>
        <Col span={6}>
          <Form.Item
            label="规则组"
            field="groupId"
            disabled={!!initialValues?.groupId}
            rules={[
              {
                required: true,
                message: "规则组不能为空, 请选择规则组",
              },
            ]}
          >
            <GroupSelect />
          </Form.Item>
        </Col>
        <Col span={18}>
          <Form.Item
            label="告警页面"
            field="alarmPageIds"
            rules={[
              {
                required: true,
                message: "规则组不能为空, 请选择告警页面",
              },
            ]}
          >
            <Select
              placeholder="请选择告警页面"
              mode="multiple"
              options={[
                ...promAlarmPages.map((item) => ({
                  label: (
                    <Tag size="small" color={item.color}>
                      {item.name}
                    </Tag>
                  ),
                  value: item.id,
                })),
              ]}
            />
          </Form.Item>
        </Col>
        <Col span={6}>
          <Form.Item
            label="告警等级"
            field="alertLevelId"
            rules={[
              {
                required: true,
                message: "规则组不能为空, 请选择规则组",
              },
            ]}
          >
            <Select
              placeholder="请选择告警等级"
              options={[
                ...promAlertLeves.map((item) => ({
                  label: (
                    <Tag size="small" color={item.color}>
                      {item.name}
                    </Tag>
                  ),
                  value: item.id,
                })),
              ]}
            />
          </Form.Item>
        </Col>
        <Col span={18}>
          <Form.Item label="规则类型" field="categorieIds">
            <Select
              placeholder="请选择规则类型"
              mode="multiple"
              options={[
                ...promCategories.map((item) => ({
                  label: (
                    <Tag size="small" color={item.color}>
                      {item.name}
                    </Tag>
                  ),
                  value: item.id,
                })),
              ]}
            />
          </Form.Item>
        </Col>
      </Row>
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
                  label={`${item.key}`}
                  rules={[
                    {
                      required: true,
                      message: `${item.key}不能为空, 请填写${item.key}`,
                    },
                  ]}
                >
                  <Input
                    size="large"
                    placeholder={`请输入${item.key}`}
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
          <Col span={24}>
            <Form.Item
              field="annotations.title"
              label="标题模板"
              rules={[{ required: true, message: "请输入报警标题模板..." }]}
            >
              <Input placeholder="请输入报警标题模板" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item
              field="annotations.remark"
              label="描述模板"
              rules={[{ required: true, message: "请输入报警描述模板..." }]}
            >
              <Input.TextArea placeholder="请输入报警描述模板" showWordLimit />
            </Form.Item>
          </Col>
          {annotations
            .filter((item) => item.key !== "remark" && item.key !== "title")
            .map((item, index) => {
              return (
                <Col span={12} key={index}>
                  <Form.Item
                    layout="vertical"
                    key={index}
                    field={`annotations.${item.key}`}
                    label={`${item.key}`}
                    rules={[
                      {
                        required: true,
                        message: `${item.key}不能为空, 请填写${item.key}`,
                      },
                    ]}
                  >
                    <Input
                      size="large"
                      placeholder={`请输入${item.key}`}
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
