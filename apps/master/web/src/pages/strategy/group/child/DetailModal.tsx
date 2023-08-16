import React from "react";
import type {GroupItem} from "@/apis/prom/prom";
import {Modal} from "@arco-design/web-react";
import groupStyle from "../style/group.module.less";
import StrategyModal from "@/pages/strategy/child/StrategyModal";

export interface DetailModalProps {
    item?: GroupItem
    children?: React.ReactNode
}

const pathPrefix = "http://localhost:9090";

const DetailModal: React.FC<DetailModalProps> = (props) => {
    const {item, children} = props;
    const [visible, setVisible] = React.useState(false);

    const handleOpenModal = () => {
        setVisible(true);
    }

    const handleCloseModal = () => {
        setVisible(false);
    }

    return (
        <>
            <div onClick={handleOpenModal}>{children}</div>
            <Modal
                visible={visible}
                onCancel={handleCloseModal}
                escToExit
                footer={null}
                className={groupStyle.GroupDetail}
                style={{width: "100%"}}
                title={item?.name}
            >
                <StrategyModal
                    title="添加策略"
                    btnProps={{
                        size: "mini",
                    }}
                    initialValues={{
                        datasource: pathPrefix,
                        alert: "up == 0",
                        expr: "up == 0",
                        for: "30s",
                        labels: {
                            severity: "critical",
                        },
                        annotations: {
                            title: "{{$labels.instance}} down",
                            description:
                                "The instance {{$labels.instance}} has been down for more than 30 seconds.",
                        },
                    }}
                />
            </Modal>
        </>
    )
}

export default DetailModal;