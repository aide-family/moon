import React, { useEffect, useState } from "react";
import { Form, Modal } from "@arco-design/web-react";
import type { Response } from "@/apis/type";

export interface AlarmModalProps {
  visible?: boolean;
  setVisible?: React.Dispatch<React.SetStateAction<boolean>>;
  initValues?: any;
  onOk?: (values: any) => Promise<Response>;
}

const AlarmModal: React.FC<AlarmModalProps> = (props) => {
  const { visible, setVisible, initValues, onOk } = props;
  const [form] = Form.useForm();

  const [loading, setLoading] = useState<boolean>(false);

  const handleOnCancel = () => {
    setVisible?.(false);
  };

  const handleOnOK = () => {
    form.validate().then((data) => {
      setLoading(true);
      onOk?.(data)
        .then((resp) => {
          handleOnCancel();
          return resp;
        })
        .finally(() => setLoading(false));
    });
  };

  useEffect(() => {
    if (form && visible) {
      form.setFieldsValue(initValues);
    }
    if (!visible) {
      form.clearFields();
    }
  }, [initValues, visible, form]);

  return (
    <>
      <Modal
        visible={visible}
        onCancel={handleOnCancel}
        onOk={handleOnOK}
        confirmLoading={loading}
      >
        <Form form={form}></Form>
      </Modal>
    </>
  );
};

export default AlarmModal;
