import React, { useEffect, useState } from "react";
import StrategyModal from "@/pages/strategy/child/StrategyModal";
import { ColumnProps } from "@arco-design/web-react/es/Table";
import { TacerTable } from "tacer-cloud";
import "./style/index.less";

export type DataSourceType = {
  id: number;
  name: string;
  group: string;
  expr: string;
  labels: { [key: string]: string };
  annotations: { [key: string]: string };
  for: string;
  datasource: string;
  // 优先级
  priority?: number;
};

const pathPrefix = "http://localhost:9090";

const Strategy: React.FC = () => {
  const tableColumns: ColumnProps<any>[] = [
    {
      title: "名称",
      dataIndex: "name",
      width: 200,
    },
    {
      title: "规则组",
      dataIndex: "group",
      width: 200,
    },
    {
      title: "表达式",
      dataIndex: "expr",
      width: 400,
    },
    {
      title: "标签",
      dataIndex: "labels",
      width: 400,
      render: (labels: { [key: string]: string }) => {
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
      render: (annotations: { [key: string]: string }) => {
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
    },
    {
      title: "数据源",
      dataIndex: "datasource",
      width: 200,
    },
    {
      title: "优先级",
      dataIndex: "priority",
      width: 200,
    },
  ];

  const [dataSource, setDataSource] = useState<DataSourceType[]>([]);

  useEffect(() => {
    let data: DataSourceType[] = [];
    for (let i = 0; i < 100; i++) {
      data.push({
        id: i,
        name: `test${i}`,
        group: `test-group${i}`,
        expr: `up == ${i}`,
        labels: {
          test: `test-${i}`,
        },
        annotations: {
          title: `title-${i}`,
          description: `description-${i}`,
        },
        for: `${i}s`,
        datasource: `datasource-${i}`,
        priority: i % 6,
      });
    }
    setDataSource(data);
  }, []);
  return (
    <div>
      <TacerTable
        rowClassName={(row) => {
          return `priority-${row.priority}`;
        }}
        rowKey={(row) => row.id}
        columns={tableColumns}
        data={dataSource}
        handleBatchExport={() => {
          return Promise.resolve();
        }}
        handleBatchDelete={() => {
          return Promise.resolve();
        }}
        columnOptionWidth={200}
        size="mini"
        showAdd={false}
        searchOptions={[
          (index: number) => {
            return (
              <StrategyModal
                title="添加策略"
                key={index}
                btnProps={{
                  size: "mini",
                }}
                initialValues={{
                  datasource: pathPrefix,
                  alert: "up == 0",
                  expr: "up == 0",
                  for: "30s",
                  labels: {
                    severity: "critical",
                  },
                  annotations: {
                    title: "{{$labels.instance}} down",
                    description:
                      "The instance {{$labels.instance}} has been down for more than 30 seconds.",
                  },
                }}
              />
            );
          },
        ]}
      />
    </div>
  );
};

export default Strategy;
