import React, { useEffect } from "react";
import { TacerTable } from "tacer-cloud";
import { ColumnProps } from "@arco-design/web-react/es/Table";
import dayjs from "dayjs";
import { TacerFormColumn } from "tacer-cloud/es/TacerForm";

export type GroupDataSourceType = {
  id: number;
  name: string;
  // 作用范围
  scope: string[];
  // 作用对象
  target: "node" | "pod" | "container" | "service" | "job" | "instance";
  creator: string;
  remark: string;
  created_at: number;
};

const Group: React.FC = () => {
  const [dataSource, setDataSource] = React.useState<GroupDataSourceType[]>([]);
  const tableColumns: ColumnProps<GroupDataSourceType>[] = [
    {
      title: "名称",
      dataIndex: "name",
      width: 200,
    },
    {
      title: "作用对象",
      dataIndex: "target",
      width: 200,
      render: (target: string) => {
        switch (target) {
          case "node":
            return "节点";
          case "pod":
            return "Pod";
          case "container":
            return "容器";
          case "service":
            return "服务";
          case "job":
            return "任务";
          case "instance":
            return "实例";
          default:
            return "未知";
        }
      },
    },
    {
      title: "作用范围",
      dataIndex: "scope",
      width: 400,
      render: (scope: string[]) => {
        return (
          <div>
            {scope.map((item) => {
              return <div key={item}>{item}</div>;
            })}
          </div>
        );
      },
    },

    {
      title: "创建人",
      dataIndex: "creator",
      width: 140,
    },
    {
      title: "创建时间",
      dataIndex: "created_at",
      width: 200,
      render: (created_at: number) => {
        return dayjs(created_at).format("YYYY-MM-DD HH:mm:ss");
      },
    },
    {
      title: "描述",
      dataIndex: "remark",
      width: 500,
    },
  ];

  const searchColumns: TacerFormColumn[] = [
    {
      type: "input",
      label: "名称",
      field: "name",
      placeholder: "请输入名称",
    },
    {
      type: "select",
      label: "作用对象",
      field: "target",
      placeholder: "请选择作用对象",
      width: 200,
      options: [
        {
          label: "节点",
          value: "node",
        },
        {
          label: "Pod",
          value: "pod",
        },
        {
          label: "容器",
          value: "container",
        },
        {
          label: "服务",
          value: "service",
        },
        {
          label: "任务",
          value: "job",
        },
      ],
    },
  ];

  useEffect(() => {
    let data: GroupDataSourceType[] = [];
    for (let i = 0; i < 100; i++) {
      data.push({
        id: i,
        name: `group-${i}`,
        scope: ["node", "pod"],
        target: "pod",
        creator: "admin",
        remark: "备注",
        created_at: 1622131200000,
      });
    }
    setDataSource(data);
  }, []);

  return (
    <>
      <TacerTable
        rowKey={(row) => row.id}
        columns={tableColumns}
        searchColumns={searchColumns}
        data={dataSource}
        size="small"
      />
    </>
  );
};

export default Group;
