import React, { useEffect } from "react";
import type { PaginationProps } from "@arco-design/web-react";
import { Button, Statistic, Table, Tag } from "@arco-design/web-react";
import type { GroupItem } from "@/apis/prom/prom";
import { PromDict, Status } from "@/apis/prom/prom";
import type { ColumnProps } from "@arco-design/web-react/es/Table";
import StatusTag from "@/pages/strategy/group/child/StatusTag";
import { calcColor, colors } from "@/utils/calcColor";
import dayjs from "dayjs";
import { GroupList } from "@/apis/prom/group/api";
import { ListGroupRequest } from "@/apis/prom/group/group";
import type { M, Sort } from "@/apis/type";
import { defaultPage } from "@/apis/type";
import { useSearchParams } from "react-router-dom";
import groupStyle from "../style/group.module.less";
import toSnakeCase from "@/utils/strings";
import MoreMenu from "@/components/More/MoreMenu";
import AddGroup from "@/pages/strategy/group/child/AddGroup";
import { OmitText } from "tacer-cloud";
import DeleteButton from "@/pages/strategy/group/child/DeleteButton";
import type { SorterInfo } from "@arco-design/web-react/es/Table/interface";
import DetailModal from "@/pages/strategy/group/child/DetailModal";

export type sizeType = "default" | "middle" | "small" | "mini";

export interface ShowTableProps {
  setQueryParams: React.Dispatch<
    React.SetStateAction<ListGroupRequest | undefined>
  >;
  queryParams?: ListGroupRequest;
  size?: sizeType;
  refresh?: boolean;
}

const ShowTable: React.FC<ShowTableProps> = (props) => {
  const { setQueryParams, queryParams, size = "default", refresh } = props;
  const [_, setSearchParams] = useSearchParams();

  const [dataSource, setDataSource] = React.useState<GroupItem[]>([]);
  const [tableLoading, setTableLoading] = React.useState<boolean>(false);
  const [tablePagination, setTablePagination] = React.useState({
    current: queryParams?.query?.page.current || defaultPage.current,
    pageSize: queryParams?.query?.page.size || defaultPage.size,
    total: 0,
  });
  const [refreshLock, setRefreshLock] = React.useState<boolean>(false);

  // 统一查询
  function onSearch() {
    if (!queryParams) return;
    setTableLoading(true);
    GroupList(queryParams || { query: { page: defaultPage } })
      .then((listGroupReply) => {
        setDataSource(() => listGroupReply.groups || []);
        let total = +listGroupReply?.result?.page?.total || 0;
        let page = total
          ? {
              total: total,
              pageSize:
                +listGroupReply?.result?.page?.size || tablePagination.pageSize,
              current:
                +listGroupReply?.result?.page?.current ||
                tablePagination.current,
            }
          : (function () {
              return defaultPage;
            })();
        setTablePagination((prev) => {
          return { ...prev, ...page };
        });
      })
      .finally(() => {
        setTableLoading(false);
      });
  }

  const tableColumns: ColumnProps<GroupItem>[] = [
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
      title: "ID",
      dataIndex: "id",
      width: 140,
      sorter: (a, b) => +a.id - +b.id,
    },
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
        return (
          <StatusTag
            onFinished={onSearch}
            status={status}
            id={item.id}
            name={item.name}
          />
        );
      },
    },
    {
      title: "规则总数",
      dataIndex: "strategyCount",
      width: 200,
      align: "center",
      render: (strategyCount: string) => {
        let color = calcColor(colors, +strategyCount, 100);
        return (
          <Statistic
            value={strategyCount}
            groupSeparator
            suffix="条"
            styleValue={{ color: color }}
          />
        );
      },
      sorter: (a, b) => +a.strategyCount - +b.strategyCount,
    },
    {
      title: "标签",
      dataIndex: "categories",
      width: 200,
      render: (categories: PromDict[]) => {
        return (
          <div style={{ gap: 8, display: "flex", flexWrap: "wrap" }}>
            {categories.map((item, index) => (
              <Tag key={index} color={item.color || "rgb(var(--arcoblue-5))"}>
                {item.name}
              </Tag>
            ))}
          </div>
        );
      },
    },
    {
      title: "描述",
      dataIndex: "remark",
      width: 500,
      render: (remark: string) => {
        return (
          <OmitText showTooltip maxLine={2} placeholder="-">
            {remark}
          </OmitText>
        );
      },
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 200,
      render: (createdAt: string) => {
        return dayjs(+createdAt * 1000).format("YYYY-MM-DD HH:mm:ss");
      },
      sorter: (a, b) => +a.createdAt - +b.createdAt,
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 200,
      render: (updatedAt: string) => {
        return dayjs(+updatedAt * 1000).format("YYYY-MM-DD HH:mm:ss");
      },
      sorter: (a, b) => +a.updatedAt - +b.updatedAt,
    },
    {
      title: "操作",
      dataIndex: "action",
      width: 120,
      fixed: "right",
      align: "center",
      render: (_, item: GroupItem) => {
        return (
          <div className={groupStyle.action}>
            <DetailModal item={item} setRefreshLock={setRefreshLock}>
              <Button type="text">详情</Button>
            </DetailModal>
            <MoreMenu
              options={[
                {
                  label: (
                    <AddGroup
                      setRefreshLock={setRefreshLock}
                      onFinished={onSearch}
                      title="编辑分组"
                      groupId={item.id}
                      initialValues={{
                        name: item.name,
                        remark: item.remark,
                        categoriesIds: item.categoriesIds,
                      }}
                    >
                      <Button type="text">编辑</Button>
                    </AddGroup>
                  ),
                  key: "edit",
                },
                {
                  label: <DeleteButton onFinished={onSearch} item={item} />,
                  key: "delete",
                },
              ]}
            />
          </div>
        );
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
  };

  useEffect(() => {
    if (!queryParams || refreshLock) return;
    setSearchParams({ q: JSON.stringify(queryParams) });
    onSearch();
  }, [queryParams, refresh]);

  const pagination: boolean | PaginationProps = {
    ...tablePagination,
    showTotal: (total) => `共 ${total} 条`,
    onChange: handlePageChange,
    sizeOptions: [10, 20, 50, 100],
    showJumper: true,
    sizeCanChange: true,
  };

  const [ShowTableDivHeight, setShowTableDivHeight] = React.useState<number>(
    document.getElementById("ShowTableDiv")?.clientHeight || 0
  );

  // 监听窗口尺寸变化
  useEffect(() => {
    const resize = () => {
      const tableDivHeight =
        document.getElementById("ShowTableDiv")?.clientHeight || 0;
      setShowTableDivHeight(tableDivHeight);
    };
    window.addEventListener("resize", resize);
    return () => {
      window.removeEventListener("resize", resize);
    };
  }, []);

  // 监听ShowTableDiv渲染完成
  useEffect(() => {
    const tableDivHeight =
      document.getElementById("ShowTableDiv")?.clientHeight || 0;
    setShowTableDivHeight(tableDivHeight);
  }, []);

  const handleSetSorter = (sorts: M<Sort>) => {
    setQueryParams?.((prev?: ListGroupRequest): ListGroupRequest | any => {
      return {
        ...prev,
        query: {
          ...prev?.query,
          sort: [...Object.keys(sorts).map((key) => sorts[key])],
        },
      };
    });
  };

  const handleTableOnChange = (
    pagination: PaginationProps,
    changedSorter: SorterInfo
  ) => {
    let sorts = {};
    if (changedSorter && changedSorter.direction) {
      let field = toSnakeCase(changedSorter.field + "");
      sorts = {
        [field]: {
          asc: changedSorter.direction === "ascend",
          field: field,
        },
      };
    }

    handleSetSorter(sorts);
  };

  return (
    <div className={groupStyle.ShowTableDiv} id="ShowTableDiv">
      <Table
        size={size}
        style={{ padding: 8 }}
        rowKey={(row) => row.id}
        loading={tableLoading}
        columns={tableColumns}
        data={dataSource}
        pagination={pagination}
        scroll={{ y: ShowTableDivHeight - 112 }}
        onChange={handleTableOnChange}
      />
    </div>
  );
};

export default ShowTable;
