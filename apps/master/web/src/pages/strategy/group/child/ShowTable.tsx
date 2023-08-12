import React, {useEffect} from "react";
import type {PaginationProps} from "@arco-design/web-react";
import {Statistic, Table, Tag} from "@arco-design/web-react";
import type {GroupItem} from "@/apis/prom/prom";
import {PromDict, Status} from "@/apis/prom/prom";
import type {ColumnProps} from "@arco-design/web-react/es/Table";
import StatusTag from "@/pages/strategy/group/child/StatusTag";
import {calcColor, colors} from "@/utils/calcColor";
import dayjs from "dayjs";
import {GroupList} from "@/apis/prom/group/api";
import {ListGroupRequest} from "@/apis/prom/group/group";
import {defaultPage} from "@/apis/type";
import {useSearchParams} from "react-router-dom";
import groupStyle from "../style/group.module.less";

export interface ShowTableProps {
    setQueryParams: React.Dispatch<React.SetStateAction<ListGroupRequest | undefined>>
    setIsNoData: React.Dispatch<React.SetStateAction<boolean>>
    queryParams?: ListGroupRequest
    isNoData: boolean
}

const ShowTable: React.FC<ShowTableProps> = (props) => {
    const {
        setQueryParams,
        setIsNoData,
        queryParams,
        isNoData,
    } = props;
    const [_, setSearchParams] = useSearchParams();

    const [dataSource, setDataSource] = React.useState<GroupItem[]>([]);
    const [tableLoading, setTableLoading] = React.useState<boolean>(false);
    const [tablePagination, setTablePagination] = React.useState({
        current: queryParams?.query?.page.current || defaultPage.current,
        pageSize: queryParams?.query?.page.size || defaultPage.size,
        total: 0,
    });

    // 统一查询
    function onSearch() {
        if (!queryParams) return;
        setTableLoading(true)
        GroupList(queryParams || {query: {page: defaultPage}}).then((listGroupReply) => {
            setDataSource(() => listGroupReply.groups || [])
            let total = +listGroupReply?.result?.page?.total || 0
            let page = total ? {
                total: total,
                pageSize: +listGroupReply?.result?.page?.size || tablePagination.pageSize,
                current: +listGroupReply?.result?.page?.current || tablePagination.current,
            } : function () {
                setIsNoData(true)
                setQueryParams?.((prev) => {
                    return {
                        ...prev,
                        query: {
                            ...prev?.query,
                            page: defaultPage
                        }
                    }
                })
                return defaultPage
            }()
            setTablePagination((prev) => {
                return {...prev, ...page,}
            })
        }).finally(() => {
            setTableLoading(false)
        })
    }

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
                return <StatusTag onFinished={onSearch} status={status} id={item.id} name={item.name}/>;
            }
        },
        {
            title: "规则总数",
            dataIndex: "strategyCount",
            width: 200,
            align: "center",
            render: (strategyCount: string) => {
                let color = calcColor(colors, +strategyCount, 100)
                return <Statistic value={strategyCount} groupSeparator suffix="条" styleValue={{color: color}}/>
            }
        },
        {
            title: "标签",
            dataIndex: "categories",
            width: 200,
            render: (categories: PromDict[]) => {
                return <div style={{gap: 8, display: "flex", flexWrap: "wrap"}}>
                    {
                        categories.map((item, index) => (
                            <Tag
                                key={index}
                                color={item.color || "rgb(var(--arcoblue-5))"}
                            >
                                {item.name}
                            </Tag>
                        ))
                    }
                </div>
            },
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

    const handlePageChange = (page: number, pageSize: number) => {
        setQueryParams?.((prev) => {
            return {
                ...prev,
                query: {
                    ...prev?.query,
                    page: {
                        ...prev?.query?.page,
                        current: page,
                        size: pageSize,
                    },
                },
            };
        });
    }

    useEffect(() => {
        if (isNoData) return;
        setSearchParams({q: JSON.stringify(queryParams)})
        onSearch()
    }, [queryParams]);

    const pagination: boolean | PaginationProps = {
        ...tablePagination,
        showTotal: (total) => `共 ${total} 条`,
        onChange: handlePageChange,
    };

    return <div className={groupStyle.ShowTableDiv}>
        <Table
            style={{padding: 8}}
            rowKey={(row) => row.id}
            loading={tableLoading}
            columns={tableColumns}
            data={dataSource}
            pagination={pagination}
        />
    </div>
}

export default ShowTable;