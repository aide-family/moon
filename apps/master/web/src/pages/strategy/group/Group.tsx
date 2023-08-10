import React, {useEffect} from "react";
import dayjs from "dayjs";
import {TacerTable} from "tacer-cloud";
import type {ColumnProps} from "@arco-design/web-react/es/Table";
import type {TacerFormColumn} from "tacer-cloud/es/TacerForm";
import type {GroupItem} from "@/apis/prom/prom";
import {Status} from "@/apis/prom/prom";
import {GroupList} from "@/apis/prom/group/api";
import {Statistic} from "@arco-design/web-react";
import {calcColor, colors} from "@/utils/calcColor";
import StatusTag from "@/pages/strategy/group/child/StatusTag";

const Group: React.FC = () => {
    const [dataSource, setDataSource] = React.useState<GroupItem[]>([]);
    const tableColumns: ColumnProps<GroupItem>[] = [
        {
            title: "名称",
            dataIndex: "name",
            width: 300,
        },
        {
            title: "状态",
            dataIndex: "status",
            width: 160,
            align: "center",
            render: (status: Status, item: GroupItem) => {
                return <StatusTag status={status} id={item.id} name={item.name}/>;
            }
        },
        {
            title: "规则总数",
            dataIndex: "strategyCount",
            width: 500,
            render: (strategyCount: string) => {
                let color = calcColor(colors, +strategyCount, 100)
                return <Statistic value={strategyCount} groupSeparator suffix="条" styleValue={{color: color}}/>
            }
        },
        {
            title: "描述",
            dataIndex: "remark",
            width: 500,
            render: (remark: string) => {
                return remark || "-";
            }
        },
        {
            title: "创建时间",
            dataIndex: "createdAt",
            width: 200,
            render: (createdAt: string) => {
                return dayjs(+createdAt).format("YYYY-MM-DD HH:mm:ss");
            },
        },
        {
            title: "更新时间",
            dataIndex: "updatedAt",
            width: 200,
            render: (updatedAt: string) => {
                return dayjs(+updatedAt).format("YYYY-MM-DD HH:mm:ss");
            },
        },
    ];

    const searchColumns: TacerFormColumn[] = [
        {
            type: "input",
            label: "名称",
            field: "name",
            placeholder: "请输入名称",
        },
    ];


    useEffect(() => {
        GroupList({
            group: undefined,
            query: {
                page: {
                    current: 1,
                    size: 1000
                }
            }
        }).then((listGroupReply) => setDataSource(() => listGroupReply.groups || []))
    }, []);

    return (
        <>
            <TacerTable
                rowKey={(row) => row.id}
                columns={tableColumns}
                searchColumns={searchColumns}
                data={dataSource}
                size="small"
                page={{
                    pageSize: 1000,
                    total: dataSource.length,
                    current: 1,
                }}
            />
        </>
    );
};

export default Group;
