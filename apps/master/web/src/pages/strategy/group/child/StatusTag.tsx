import React, {createContext} from "react";

import {Button, Modal, Tag} from "@arco-design/web-react";
import {Tooltip} from "@arco-design/web-react/lib";
import {ConfirmProps} from "@arco-design/web-react/es/Modal/confirm";
import Countdown from "@arco-design/web-react/es/Statistic/countdown";

import dayjs from "dayjs";

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

const defaultTime = 3;

const StatusTag: React.FC<StatusTagProps> = (props) => {
    const {status, name, id, onFinished} = props;
    const statusValue = StatusMap[status];

    const [modal, contextHolder] = Modal.useModal();

    const Footer = () => {
        const [loading, setLoading] = React.useState<boolean>(false);
        const [closeLoading, setCloseLoading] = React.useState<boolean>(false);
        const init = () => {
            setLoading(false)
            setCloseLoading(false)
        }

        return <div style={{flex: "1 1 100%", display: "flex", justifyContent: "flex-end", gap: "12px"}}>
            <Button type="secondary" loading={closeLoading} onClick={() => {
                setCloseLoading(true)
                setTimeout(() => {
                    init()
                    Modal.destroyAll()
                }, loading ? 500 : 0)
            }}>取消</Button>
            <Button
                status={status !== Status.Status_ENABLE ? "default" : "danger"}
                type="primary"
                loading={loading || closeLoading}
                onClick={() => {
                    setLoading(true);
                }}
            >{statusValue.opposite?.text}{" "}{loading &&
                <span>
                    <Countdown
                        format="ss"
                        styleValue={{color: "var(--color-neutral-8)", fontSize: "16px", fontWeight: "bold"}}
                        value={dayjs().add(defaultTime, "second")}
                        start={loading}
                        now={dayjs()}
                        onFinish={
                            () => {
                                GroupUpdatesStatus([id], status).then(onFinished).finally(() => {
                                    setLoading(false)
                                    Modal.destroyAll()
                                })
                            }
                        }/>
                    s
                </span>
            }
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