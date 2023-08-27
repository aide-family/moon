import React, { useEffect, useState } from "react";
import { ListStrategyRequest } from "@/apis/prom/strategy/strategy";
import {
  StrategyDelete,
  StrategyList,
  StrategyUpdate,
} from "@/apis/prom/strategy/api";
import {
  Button,
  Message,
  PaginationProps,
  Table,
  Tag,
} from "@arco-design/web-react";
import { ColumnProps } from "@arco-design/web-react/es/Table";
import {
  AlarmPage,
  PromDict,
  PromGroupSimpleItem,
  PromStrategyItem,
  Status,
} from "@/apis/prom/prom";
import { defaultPage, M } from "@/apis/type";
import MoreMenu from "@/components/More/MoreMenu";
import StatusTag from "@/pages/strategy/child/StatusTag";
import { OmitText } from "tacer-cloud";
import {
  StrategyModalProps,
  StrategyValues,
} from "@/pages/strategy/child/StrategyModal";

import strategyStyles from "../style/strategy.module.less";
import groupStyle from "@/pages/strategy/group/style/group.module.less";

export type tableSize = "default" | "middle" | "small" | "mini";

export interface ShowTableProps {
  size?: tableSize;
  className?: string;
  database?: string;
  queryParams?: ListStrategyRequest;
  setQueryParams?: React.Dispatch<React.SetStateAction<ListStrategyRequest>>;
  refresh?: boolean;
  setRefresh?: React.Dispatch<React.SetStateAction<boolean>>;
  setStrategyModalProps?: React.Dispatch<
    React.SetStateAction<StrategyModalProps | undefined>
  >;
  setStrategyModalVisabled?: React.Dispatch<React.SetStateAction<boolean>>;
}

const ShowTable: React.FC<ShowTableProps> = (props) => {
  const {
    size,
    className,
    queryParams,
    setQueryParams,
    database,
    refresh,
    setRefresh,
    setStrategyModalProps,
    setStrategyModalVisabled,
  } = props;
  const [tableLoading, setTableLoading] = useState<boolean>(false);

  const updateStrategy = async (
    val: StrategyValues,
    item: PromStrategyItem
  ) => {
    const res = await StrategyUpdate(item.id, {
      alarmPageIds: val.alarmPageIds || item.alarmPageIds,
      alert: val.alert || item.alert,
      alertLevelId: val.alertLevelId || item.alertLevelId,
      annotations: val.annotations || item.annotations,
      categorieIds: val.categorieIds || item.categorieIds,
      expr: val.expr || item.expr,
      for: val.for || item.for,
      groupId: val.groupId || item.groupId,
      labels: val.labels || item.labels,
    });
    if (res.response.code !== "0") {
      Message.error(res.response.message);
      return Promise.reject(res.response.message);
    }
    if (setRefresh) {
      setRefresh?.((prv) => !prv);
    } else {
      getStrategies();
    }

    return res;
  };

  // 删除告警规则
  const handleDeleteStrategy = (id: number) => {
    StrategyDelete(id).then(() => {
      setRefresh?.((prev) => !prev);
    });
  };

  const tableColumns: ColumnProps<PromStrategyItem>[] = [
    {
      title: "序号",
      dataIndex: "index",
      width: 60,
      fixed: "left",
      align: "left",
      render: (text, item, index) => {
        return (
          <span>
            {(tablePagination.current - 1) * tablePagination.pageSize +
              index +
              1}
          </span>
        );
      },
    },
    {
      title: "规则组",
      dataIndex: "group",
      width: 200,
      align: "center",
      render: (group: PromGroupSimpleItem) => {
        return <Button type="text">{group.name}</Button>;
      },
    },
    // {
    //   title: "id",
    //   dataIndex: "id",
    //   align: "left",
    //   width: 120,
    // },
    {
      title: "名称",
      dataIndex: "alert",
      align: "left",
      width: 200,
      render: (_, row) => {
        return (
          <OmitText
            showTooltip
            maxLine={2}
          >{`[${row.id}] ${row.alert}`}</OmitText>
        );
      },
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 160,
      align: "center",
      render: (status: Status, item: PromStrategyItem) => {
        return (
          <StatusTag
            onFinished={getStrategies}
            status={status}
            id={item.id}
            name={item.alert}
          />
        );
      },
    },
    {
      title: "报警页面",
      dataIndex: "alarmPages",
      width: 300,
      render: (alarmPages: AlarmPage[]) => {
        return (
          <div className={strategyStyles.alarmPagesDiv}>
            {alarmPages.map((alarmPage, index) => {
              return (
                <Tag key={index} color={alarmPage.color}>
                  {alarmPage.remark}
                </Tag>
              );
            })}
          </div>
        );
      },
    },
    // {
    //   title: "标签",
    //   dataIndex: "labels",
    //   width: 400,
    //   render: (labels: M) => {
    //     if (!labels) return "-";
    //     return Object.keys(labels).map((key) => {
    //       return <div key={key}>{`${key}(${labels[key]})`}</div>;
    //     });
    //   },
    // },
    // {
    //   title: "注释",
    //   dataIndex: "annotations",
    //   width: 400,
    //   render: (annotations: M) => {
    //     if (!annotations) return "-";
    //     return Object.keys(annotations).map((key) => {
    //       return <div key={key}>{`${key}(${annotations[key]})`}</div>;
    //     });
    //   },
    // },
    {
      title: "持续时间",
      dataIndex: "for",
      width: 120,
      align: "center",
    },
    {
      title: "优先级",
      dataIndex: "alertLevel",
      width: 160,
      align: "center",
      render: (alertLevel: PromDict) => {
        return <Tag color={alertLevel.color}>{alertLevel.name}</Tag>;
      },
    },
    {
      title: "规则属性",
      dataIndex: "categories",
      width: 160,
      align: "center",
      render: (categories: PromDict[]) => {
        return (
          <div className={strategyStyles.alarmPagesDiv}>
            {categories.map((categorie, index) => {
              return (
                <Tag key={index} color={categorie.color}>
                  {categorie.remark}
                </Tag>
              );
            })}
          </div>
        );
      },
    },
    {
      title: "表达式",
      dataIndex: "expr",
      width: 400,
      render: (expr: string) => {
        return (
          <OmitText showTooltip maxLine={2} placeholder="-">
            {expr}
          </OmitText>
        );
      },
    },
    {
      title: "操作",
      dataIndex: "action",
      width: 120,
      fixed: "right",
      align: "center",
      render: (_, row) => {
        return (
          <div className={groupStyle.action} key={row.id}>
            <Button
              type="text"
              onClick={() => {
                setStrategyModalVisabled?.(true);
                setStrategyModalProps?.({
                  disabled: true,
                  title: "详情",
                  setVisible: setStrategyModalVisabled,
                  initialValues: {
                    datasource: database,
                    groupId: row.groupId,
                    alertLevelId: row.alertLevelId,
                    alarmPageIds: row.alarmPageIds,
                    categorieIds: row.categorieIds,
                    alert: row.alert,
                    for: row.for,
                    expr: row.expr,
                    labels: row.labels,
                    annotations: row.annotations,
                  },
                });
              }}
            >
              详情
            </Button>

            <MoreMenu
              options={[
                {
                  label: (
                    <Button
                      type="text"
                      onClick={() => {
                        setStrategyModalVisabled?.(() => {
                          setStrategyModalProps?.({
                            onOk: (newVal) => updateStrategy(newVal, row),
                            title: "编辑",
                            setVisible: setStrategyModalVisabled,
                            initialValues: {
                              datasource: database,
                              groupId: row.groupId,
                              alertLevelId: row.alertLevelId,
                              categorieIds: row.categorieIds,
                              alarmPageIds: row.alarmPageIds,
                              alert: row.alert,
                              for: row.for,
                              expr: row.expr,
                              labels: row.labels,
                              annotations: row.annotations,
                            },
                          });
                          return true;
                        });
                      }}
                    >
                      编辑
                    </Button>
                  ),
                  key: "eidt",
                },
                {
                  label: (
                    <Button
                      status="danger"
                      type="text"
                      onClick={() => handleDeleteStrategy(row.id)}
                    >
                      删除
                    </Button>
                  ),
                  key: "delete",
                },
              ]}
            />
          </div>
        );
      },
    },
  ];

  const [dataSource, setDataSource] = useState<PromStrategyItem[]>([]);
  const [tablePagination, setTablePagination] = React.useState({
    current: queryParams?.query?.page.current || defaultPage.current,
    pageSize: queryParams?.query?.page.size || defaultPage.size,
    total: 0,
  });

  const getStrategies = () => {
    if (!queryParams) return;
    setTableLoading(true);
    StrategyList(queryParams)
      .then((data) => {
        const {
          strategies,
          result: {
            page: { current, size, total },
          },
        } = data;
        setDataSource(strategies || []);
        setTablePagination({ current: current, pageSize: size, total: +total });
      })
      .finally(() => setTableLoading(false));
  };

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
  };

  const pagination: boolean | PaginationProps = {
    ...tablePagination,
    showTotal: (total) => `共 ${total} 条`,
    onChange: handlePageChange,
    sizeOptions: [10, 20, 50, 100],
    showJumper: true,
    sizeCanChange: true,
  };

  useEffect(() => {
    setDataSource([]);
    if (!queryParams) return;
    getStrategies();
  }, [queryParams, refresh]);

  const [ShowTableDivHeight, setShowTableDivHeight] = React.useState<number>(
    document.getElementById("strategyTabl")?.clientHeight || 0
  );

  // 监听窗口尺寸变化
  useEffect(() => {
    const resize = () => {
      const tableDivHeight =
        document.getElementById("strategyTabl")?.clientHeight || 0;
      setShowTableDivHeight(tableDivHeight);
    };
    window.addEventListener("resize", resize);
    return () => {
      window.removeEventListener("resize", resize);
    };
  }, []);

  return (
    <>
      <div id="strategyTabl" className={className}>
        <Table
          size={size}
          rowKey={(record) => record.id}
          loading={tableLoading}
          data={dataSource}
          columns={tableColumns}
          pagination={pagination}
          scroll={{ y: ShowTableDivHeight - 112 }}
        />
      </div>
    </>
  );
};

export default ShowTable;
