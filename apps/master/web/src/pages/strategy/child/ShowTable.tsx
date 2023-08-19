import React, { useEffect, useState } from "react";
import { ListStrategyRequest } from "@/apis/prom/strategy/strategy";
import { StrategyList } from "@/apis/prom/strategy/api";
import { Button, PaginationProps, Table, Tag } from "@arco-design/web-react";
import { ColumnProps } from "@arco-design/web-react/es/Table";
import {
  AlarmPage,
  PromDict,
  PromGroupSimpleItem,
  PromStrategyItem,
} from "@/apis/prom/prom";
import { defaultPage, M } from "@/apis/type";

import strategyStyles from "../style/strategy.module.less";
import { OmitText } from "tacer-cloud";
import StrategyModal from "@/pages/strategy/child/StrategyModal";
import groupStyle from "@/pages/strategy/group/style/group.module.less";
import MoreMenu from "@/components/More/MoreMenu";

export interface ShowTableProps {
  database?: string;
  queryParams?: ListStrategyRequest;
  setQueryParams?: React.Dispatch<React.SetStateAction<ListStrategyRequest>>;
}

const ShowTable: React.FC<ShowTableProps> = (props) => {
  const { queryParams, setQueryParams, database } = props;
  const [tableLoading, setTableLoading] = React.useState<boolean>(false);

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
      title: "操作",
      dataIndex: "action",
      width: 120,
      fixed: "right",
      render: (_, row) => {
        return (
          <div className={groupStyle.action} key={row.id}>
            <StrategyModal
              disabled
              title="详情"
              btnProps={{ type: "text" }}
              key={row.id}
              initialValues={{
                datasource: database,
                alert: row.alert,
                for: row.for,
                expr: row.expr,
                labels: row.labels,
                annotations: row.annotations,
              }}
            />
            <MoreMenu
              options={[
                {
                  label: (
                    <StrategyModal
                      title="编辑"
                      btnProps={{ type: "text" }}
                      initialValues={{
                        datasource: database,
                        alert: row.alert,
                        for: row.for,
                        expr: row.expr,
                        labels: row.labels,
                        annotations: row.annotations,
                      }}
                    />
                  ),
                  key: "eidt",
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
  }, [queryParams]);

  return (
    <Table
      rowKey={(record) => record.id}
      loading={tableLoading}
      data={dataSource}
      columns={tableColumns}
      pagination={pagination}
    />
  );
};

export default ShowTable;
