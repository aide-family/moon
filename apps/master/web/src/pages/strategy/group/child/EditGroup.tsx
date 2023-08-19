import React, { useEffect } from "react";
import { Form, Input, Modal, Select } from "@arco-design/web-react";
import type { GroupItem } from "@/apis/prom/prom";
import { Response } from "@/apis/type";
import { GroupCreate, GroupUpdate } from "@/apis/prom/group/api";
import { GroupCreateItem } from "@/apis/prom/group/group";

export interface EditGroupProps {
  title: React.ReactNode;
  onFinished?: () => void;
  visible?: boolean;
  onClose?: () => void;
  initialValues?: GroupCreateItem;
  id?: number;
}

const EditGroup: React.FC<EditGroupProps> = (props) => {
  const { onFinished, visible, onClose, initialValues, id } = props;

  const [form] = Form.useForm();

  const handleOnOk = () => {
    form
      .validate()
      .then((values: GroupItem) => {
        if (initialValues && id) {
          return onEditGroup(id, values);
        }
        return onAddGroup(values);
      })
      .then(onFinished);
  };

  const onAddGroup = (data: GroupCreateItem): Promise<Response> => {
    return GroupCreate(data);
  };

  const onEditGroup = (
    id: number,
    data: GroupCreateItem
  ): Promise<Response> => {
    return GroupUpdate(id, data);
  };

  useEffect(() => {
    if (initialValues && visible) {
      form.setFieldsValue(initialValues);
    }
    if (!visible && form) {
      form.resetFields();
    }
  }, [initialValues, visible]);

  return (
    <>
      <Modal
        visible={visible}
        title={props.title}
        onOk={handleOnOk}
        onCancel={onClose}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="分组名称"
            field="name"
            rules={[
              {
                required: true,
                message: "请输入分组名称",
              },
              {
                match: /^[\u4e00-\u9fa5_a-zA-Z0-9]+$/,
                message: "分组名称只能包含中文、英文、数字和下划线",
              },
            ]}
          >
            <Input placeholder="请输入分组名称" autoComplete="off" />
          </Form.Item>
          <Form.Item label="标签" field="categoriesIds">
            <Select
              placeholder="请选择标签"
              mode="multiple"
              options={[]}
              showSearch
            />
          </Form.Item>
          <Form.Item label="备注" field="remark">
            <Input.TextArea
              placeholder="请输入备注"
              maxLength={255}
              showWordLimit
              autoComplete="off"
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default EditGroup;
