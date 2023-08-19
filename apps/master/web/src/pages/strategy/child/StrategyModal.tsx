import React, { useState } from "react";
import { Button, Form, Modal, Space } from "@arco-design/web-react";
import StrategyForm from "@/pages/strategy/child/StrategyForm";
import CodeView from "@/components/Code/CodeView";
import { toYaml } from "@/utils/yaml";
import type { M, Response } from "@/apis/type";

export interface StrategyValues {
  datasource?: string;
  labels?: M;
  annotations?: M;
  for?: string;
  alert?: string;
  expr?: string;
}

export interface StrategyModalProps {
  title?: string;
  initialValues?: StrategyValues;
  onOk?: (values: StrategyValues) => Promise<Response>;
  disabled?: boolean;
  visible?: boolean;
  setVisible?: React.Dispatch<React.SetStateAction<boolean>>;
}

const StrategyModal: React.FC<StrategyModalProps> = (props) => {
  const { title, disabled, initialValues, onOk, visible, setVisible } = props;
  const [form] = Form.useForm();
  const [data, setData] = useState<StrategyValues | undefined>(initialValues);
  const [loading, setLoading] = useState<boolean>(false);

  const handleOnClose = () => {
    setVisible?.(false);
  };

  const handleOnOk = () => {
    form.validate().then((val) => {
      setLoading(true);
      onOk?.(val)
        .then(handleOnClose)
        .finally(() => setLoading(false));
    });
  };

  const handleSetValues = (formValues?: StrategyValues) => {
    setVisible?.(true);
    setData(formValues);
  };

  const View = (props: { rule?: StrategyValues }) => {
    const { rule } = props;
    const [open, setOpen] = React.useState(false);

    const handleOnOpen = () => {
      setOpen(true);
    };

    const handleOnClose = () => {
      setOpen(false);
    };

    const toYamlString = (jsonData?: StrategyValues) => {
      // 去掉jsonData的datasource属性
      let { datasource, ...rest } = jsonData || {};
      // 去掉undefined的属性
      const removeUndefined = (obj: any) => {
        Object.keys(obj).forEach((key) => {
          if (obj[key] === undefined) {
            delete obj[key];
          }
          // 判断是否不是基础类型
          if (typeof obj[key] === "object" && obj[key] !== null) {
            removeUndefined(obj[key]);
          }
        });
        return obj;
      };
      return toYaml(removeUndefined(rest));
    };

    return (
      <>
        <Button
          type="text"
          size="small"
          onClick={handleOnOpen}
          style={{ float: "right" }}
        >
          预览
        </Button>
        <Modal
          unmountOnExit
          visible={open}
          title="预览"
          onCancel={handleOnClose}
          style={{ width: "60vw" }}
          maskStyle={{ padding: 0 }}
        >
          <CodeView codeString={toYamlString(rule)} language="yaml" />
        </Modal>
      </>
    );
  };

  const Title = (props: { title?: string }) => {
    return (
      <Space style={{ width: "100%" }}>
        <span style={{ float: "left" }}>{props.title}</span>
        <View rule={data} />
      </Space>
    );
  };

  return (
    <>
      <Modal
        visible={visible}
        unmountOnExit
        closable={false}
        title={<Title title={title} />}
        onCancel={handleOnClose}
        onOk={handleOnOk}
        style={{ width: "80vw" }}
        okButtonProps={{ loading, disabled }}
      >
        <div
          style={{ maxHeight: "80vh", overflow: "auto", overflowX: "hidden" }}
        >
          <div>
            <StrategyForm
              form={form}
              disabled={disabled}
              initialValues={initialValues}
              onChange={handleSetValues}
            />
          </div>
        </div>
      </Modal>
    </>
  );
};

export default StrategyModal;
