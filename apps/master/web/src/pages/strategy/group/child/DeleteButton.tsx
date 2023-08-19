import React from "react";
import { GroupItem } from "@/apis/prom/prom";
import { GroupDelete } from "@/apis/prom/group/api";
import { Button, Modal } from "@arco-design/web-react";

export interface DeleteButtonProps {
  children?: React.ReactNode;
  item?: GroupItem;
  onFinished?: () => void;
}

const DeleteButton: React.FC<DeleteButtonProps> = (props) => {
  const { children, item, onFinished } = props;
  const [visible, setVisible] = React.useState<boolean>(false);
  const [loading, setLoading] = React.useState<boolean>(false);

  const deletGroup = (id: number) => {
    setLoading(true);
    GroupDelete(id)
      .then(onFinished)
      .finally(() => setLoading(false));
  };

  const handleClick = () => {
    setVisible(true);
  };

  const handleOncancel = () => {
    setVisible(false);
  };

  const handleOnOK = () => {
    if (item) {
      deletGroup(item.id);
    }
  };

  return (
    <div>
      <div onClick={handleClick}>
        {children || (
          <Button type="text" status="danger">
            删除
          </Button>
        )}
      </div>
      <Modal
        visible={visible}
        onCancel={handleOncancel}
        onOk={handleOnOK}
        title={`删除分组 ${item?.name}确认`}
        closeIcon={null}
        okButtonProps={{ loading }}
      >
        <div>确定要删除分组 {item?.name} 吗？</div>
      </Modal>
    </div>
  );
};

export default DeleteButton;
