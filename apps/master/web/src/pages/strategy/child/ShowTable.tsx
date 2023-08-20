import React, { useEffect, useState } from "react";
import { ListStrategyRequest } from "@/apis/prom/strategy/strategy";
import { StrategyList, StrategyUpdate } from "@/apis/prom/strategy/api";
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
import StrategyModal, {
  StrategyModalProps,
  StrategyValues,
} from "@/pages/strategy/child/StrategyModal";

import strategyStyles from "../style/strategy.module.less";
import groupStyle from "@/pages/strategy/group/style/group.module.less";

export interface ShowTableProps {
  database?: string;
  queryParams?: ListStrategyRequest;
  setQueryParams?: React.Dispatch<React.SetStateAction<ListStrategyRequest>>;
  refresh?: boolean;
  setRefresh?: React.Dispatch<React.SetStateAction<boolean>>;
}

const ShowTable: React.FC<ShowTableProps> = (props) => {
  const { queryParams, setQueryParams, database, refresh, setRefresh } = props;
  const [tableLoading, setTableLoading] = useState<boolean>(false);
  const [strategyModalProps, setStrategyModalProps] = useState<
    StrategyModalProps | undefined
  >();
  const [strategyModalVisabled, setStrategyModalVisabled] =
    useState<boolean>(false);

  const updateStrategy = async (
    val: StrategyValues,
    item: PromStrategyItem
  ) => {
    const res = await StrategyUpdate(item.id, {
      alarmPageIds: item.alarmPageIds,
      alert: val.alert || item.alert,
      alertLevelId: item.alertLevelId,
      annotations: val.annotations || item.annotations,
      categorieIds: item.categorieIds,
      expr: val.expr || item.expr,
      for: val.for || item.for,
      groupId: item.groupId,
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

  const tableColumns: ColumnProps<PromStrategyItem>[] = [
    {
      title: "序号",
      dataIndex: "index",
      width: 80,
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
      title: "名称",
      dataIndex: "alert",
      width: 200,
    },
    {
      title: "规则组",
      dataIndex: "group",
      width: 200,
      render: (group: PromGroupSimpleItem) => {
        return <Button type="text">{group.name}</Button>;
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
      title: "标签",
      dataIndex: "labels",
      width: 400,
      render: (labels: M) => {
        if (!labels) return "-";
        return Object.keys(labels).map((key) => {
          return (
            <div key={key}>
              {key}: {labels[key]}
            </div>
          );
        });
      },
    },
    {
      title: "注释",
      dataIndex: "annotations",
      width: 400,
      render: (annotations: M) => {
        if (!annotations) return "-";
        return Object.keys(annotations).map((key) => {
          return (
            <div key={key}>
              {key}: {annotations[key]}
            </div>
          );
        });
      },
    },
    {
      title: "持续时间",
      dataIndex: "for",
      width: 120,
      align: "center",
    },
    {
      title: "优先级",
      dataIndex: "alertLevel",
      width: 200,
      render: (alertLevel: PromDict) => {
        return <Tag color={alertLevel.color}>{alertLevel.name}</Tag>;
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
      title: "操作",
      dataIndex: "action",
      width: 120,
      fixed: "right",
      render: (_, row) => {
        return (
          <div className={groupStyle.action} key={row.id}>
            <Button
              type="text"
              onClick={() => {
                setStrategyModalVisabled(true);
                setStrategyModalProps({
                  disabled: true,
                  title: "详情",
                  setVisible: setStrategyModalVisabled,
                  initialValues: {
                    datasource: database,
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
                        setStrategyModalVisabled(() => {
                          setStrategyModalProps({
                            onOk: (newVal) => updateStrategy(newVal, row),
                            title: "编辑",
                            setVisible: setStrategyModalVisabled,
                            initialValues: {
                              datasource: database,
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
                  label: <Button type="text">页面</Button>,
                  key: "alarm-pages",
                },
                {
                  label: <Button type="text">等级</Button>,
                  key: "alarm-level",
                },
                {
                  label: (
                    <Button status="danger" type="text">
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

  return (
    <>
      <Table
        rowKey={(record) => record.id}
        loading={tableLoading}
        data={dataSource}
        columns={tableColumns}
        pagination={pagination}
      />
      <StrategyModal {...strategyModalProps} visible={strategyModalVisabled} />
    </>
  );
};

export default ShowTable;
