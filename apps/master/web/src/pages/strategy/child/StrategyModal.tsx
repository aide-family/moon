import React, { createContext } from "react";
import { Button, Form, Modal, Space } from "@arco-design/web-react";
import type { ButtonProps } from "@arco-design/web-react/es/Button";
import StrategyForm from "@/pages/strategy/child/StrategyForm";
import CodeView from "@/components/Code/CodeView";
import { toYaml } from "@/utils/yaml";

export interface StrategyValues {
  datasource?: string;
  labels?: { [key: string]: string };
  annotations?: { [key: string]: string };
  for?: string;
  alert?: string;
  expr?: string;
}

export interface StrategyModalProps {
  title: string;
  btnProps?: ButtonProps;
  initialValues?: StrategyValues;
  onChange?: (values?: StrategyValues) => void;
}

const ConfigContext = createContext({});

const StrategyModal: React.FC<StrategyModalProps> = (props) => {
  const { title, btnProps, initialValues, onChange } = props;
  const [form] = Form.useForm();
  const [visible, setVisible] = React.useState(false);
  const [data, setData] = React.useState<StrategyValues | undefined>(
    initialValues
  );

  const handleOnOpen = () => {
    setVisible(true);
  };

  const handleOnClose = () => {
    setVisible(false);
  };

  const handleOnOk = () => {
    form.validate().then((val) => {
      console.log("val", val);
      handleOnClose();
    });
  };

  const handleSetValues = (formValues?: StrategyValues) => {
    onChange?.(formValues);
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

  const Title = (props: { title: string }) => {
    return (
      <Space style={{ width: "100%" }}>
        <span
          style={{
            float: "left",
          }}
        >
          {props.title}
        </span>
        <View rule={data} />
      </Space>
    );
  };

  return (
    <>
      <ConfigContext.Provider
        value={{
          name: "hello",
          age: 18,
        }}
      >
        <Modal
          visible={visible}
          title={<Title title={title} />}
          onCancel={handleOnClose}
          onOk={handleOnOk}
          closable={false}
          style={{
            width: "80vw",
          }}
        >
          <ConfigContext.Consumer>
            {() => {
              return (
                <div
                  style={{
                    maxHeight: "80vh",
                    overflow: "auto",
                    overflowX: "hidden",
                  }}
                >
                  <div>
                    <StrategyForm
                      form={form}
                      initialValues={initialValues}
                      onChange={handleSetValues}
                    />
                  </div>
                </div>
              );
            }}
          </ConfigContext.Consumer>
        </Modal>
        <Button type="primary" onClick={handleOnOpen} {...btnProps}>
          {title}
        </Button>
      </ConfigContext.Provider>
    </>
  );
};

export default StrategyModal;
