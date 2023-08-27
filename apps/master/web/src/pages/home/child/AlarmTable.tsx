import React from "react";
import type { AlarmItem, AlarmQueryParams } from "@/apis/prom/alarm/alarm";
import type { sizeType } from "@/pages/home/child/AlarmOption";
import { Button, PaginationProps, Table } from "@arco-design/web-react";
import type { ColumnProps } from "@arco-design/web-react/es/Table";
import { defaultPage } from "@/apis/type";

import styles from "../style/alarm.module.less";
import type { MoreMenuOption } from "@/components/More/MoreMenu";
import MoreMenu from "@/components/More/MoreMenu";

export interface AlarmTableProps {
  refresh?: () => void;
  queryParams?: AlarmQueryParams;
  size?: sizeType;
}

const AlarmTable: React.FC<AlarmTableProps> = (props) => {
  const { refresh, queryParams, size } = props;
  const [tablePagination, setTablePagination] = React.useState({
    current: queryParams?.query?.page.current || defaultPage.current,
    pageSize: queryParams?.query?.page.size || defaultPage.size,
    total: 0,
  });
  const [dataSource, setDataSource] = React.useState<AlarmItem[]>([]);
  const [tableLoading, setTableLoading] = React.useState<boolean>(false);

  const actionOptions: MoreMenuOption[] = [
    {
      key: "intervene",
      label: <Button type="text">介入</Button>,
    },
  ];

  const tableColumns: ColumnProps<AlarmItem>[] = [
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
      title: "操作",
      dataIndex: "action",
      width: 120,
      fixed: "right",
      align: "center",
      render: (_, item: AlarmItem) => {
        return (
          <div className={styles.action}>
            <Button type="primary">详情</Button>
            <MoreMenu options={actionOptions} />
          </div>
        );
      },
    },
  ];

  const handlePageChange = (page: number, pageSize: number) => {};

  const pagination: boolean | PaginationProps = {
    ...tablePagination,
    showTotal: (total) => `共 ${total} 条`,
    onChange: handlePageChange,
    sizeOptions: [10, 20, 50, 100],
    showJumper: true,
    sizeCanChange: true,
  };

  return (
    <>
      <div className={styles.alarmTable}>
        <Table
          rowKey={(row) => row.id}
          columns={tableColumns}
          pagination={pagination}
          loading={tableLoading}
          data={dataSource}
          size={size}
        />
      </div>
    </>
  );
};

export default AlarmTable;
