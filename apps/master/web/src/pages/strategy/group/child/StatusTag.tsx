import React, {createContext} from "react";

import {Button, Modal, Tag} from "@arco-design/web-react";
import {Tooltip} from "@arco-design/web-react/lib";
import {ConfirmProps} from "@arco-design/web-react/es/Modal/confirm";

import type {Response} from "@/apis/type";
import {Status, StatusMap} from "@/apis/prom/prom";
import {GroupUpdatesStatus} from "@/apis/prom/group/api";

export interface StatusTagProps {
    status: Status;
    name: string;
    id: number
    onFinished?: (resp: Response) => void
}

const ConfigContext = createContext({});

const StatusTag: React.FC<StatusTagProps> = (props) => {
    const {status, name, id, onFinished} = props;
    const statusValue = StatusMap[status];

    const [modal, contextHolder] = Modal.useModal();

    const Footer = () => {
        const [loading, setLoading] = React.useState<boolean>(false);
        const init = () => {
            setLoading(false)
        }

        return <div style={{flex: "1 1 100%", display: "flex", justifyContent: "flex-end", gap: "12px"}}>
            <Button type="secondary" onClick={() => {
                init()
                Modal.destroyAll()
            }}>取消</Button>
            <Button
                status={status !== Status.Status_ENABLE ? "default" : "danger"}
                type="primary"
                loading={loading}
                onClick={() => {
                    setLoading(true);
                    GroupUpdatesStatus([id], status).then(onFinished).finally(() => {
                        setLoading(false)
                        Modal.destroyAll()
                    })
                }}
            >
                {statusValue.opposite?.text}
            </Button>
        </div>
    }

    const config: ConfirmProps = {
        title: `${statusValue.opposite?.text}提醒`,
        content: <ConfigContext.Consumer>{(name: any) => {
            return <div style={{width: "100%", textAlign: "center"}}>确认{statusValue.opposite?.text}
                <span style={{color: "red"}}><b>{name}</b></span>规则吗?
            </div>
        }}</ConfigContext.Consumer>,
        footer: <Footer/>,
    };

    return <>
        <ConfigContext.Provider value={name}>
            {contextHolder}
            <Tooltip content="点击修改状态">
                <Tag onClick={() => {
                    if (!modal.confirm) {
                        return
                    }
                    modal.confirm(config)
                }}
                     color={statusValue.color}>{statusValue.text}</Tag>
            </Tooltip>
        </ConfigContext.Provider>
    </>
}

export default StatusTag;