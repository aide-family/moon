import React from "react";

import {Button, Form, Input, Modal} from "@arco-design/web-react";

export interface PromQLFormItemProps {
    onFinished: (value: Map) => void;
    title: string;
    disabled?: boolean
}

export type Map = {
    name: string;
    key: string;
};

const MapFormModal: React.FC<PromQLFormItemProps> = (props) => {
    const {onFinished, title, disabled} = props;
    const [form] = Form.useForm();
    const [visible, setVisible] = React.useState(false);

    const handleOnOpen = () => {
        setVisible(true);
    };

    const handleOnClose = () => {
        setVisible(false);
    };

    const handleOnOk = () => {
        form.validate().then((value) => {
            onFinished(value);
            form.clearFields();
            handleOnClose();
        });
    };

    return (
        <>
            <Button type="primary" size="mini" onClick={handleOnOpen} disabled={disabled}>
                {title}
            </Button>
            <Modal
                visible={visible}
                title="添加标签"
                onCancel={handleOnClose}
                onOk={handleOnOk}
            >
                <Form layout="vertical" autoComplete="off" form={form}>
                    <Form.Item
                        label="标签名称"
                        field="name"
                        rules={[
                            {
                                required: true,
                                message: "标签名称不能为空, 请填写标签名称",
                            },
                        ]}
                    >
                        <Input placeholder="请输入标签名称"/>
                    </Form.Item>
                    <Form.Item
                        label="标签Key"
                        field="key"
                        rules={[
                            {
                                required: true,
                                message: "标签Key不能为空, 请填写标签Key",
                            },
                            {
                                validator: (value, callback) => {
                                    // 不能是数字开头的正则
                                    const reg = /^[^0-9][\w-]*$/;
                                    if (!reg.test(value)) {
                                        callback("key不能以数字开头");
                                        return;
                                    }
                                    callback();
                                },
                            },
                        ]}
                    >
                        <Input placeholder="请输入标签Key"/>
                    </Form.Item>
                </Form>
            </Modal>
        </>
    );
};

export default MapFormModal;
