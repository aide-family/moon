import React, { useEffect } from "react";
import {
  Alert,
  Button,
  List,
  Modal,
  Space,
  Tabs,
} from "@arco-design/web-react";
import dayjs from "dayjs";
import { CopyText } from "tacer-cloud";
import SearchForm, { DateDataType } from "@/components/Prom/SearchForm";
import AreaStackGradient from "@/components/charts/area-stack-gradient/AreaStackGradient";
import { promData } from "@/components/charts/area-stack-gradient/option";
import { IconImage, IconList } from "@arco-design/web-react/icon";

export interface PromValueModalProps {
  visible: boolean;
  onCancel: () => void;
  pathPrefix: string;
  apiPath?: string;
  expr?: string;
  height?: number | string;
}

export interface PromValue {
  metric: {
    __name__: string;
    instance: string;
    [key: string]: string;
  };
  value?: [number, string];
  values?: [number, string][];
}

const TabPane = Tabs.TabPane;

const now = dayjs();

const PromValueModal: React.FC<PromValueModalProps> = ({
  visible,
  onCancel,
  pathPrefix,
  apiPath = "api/v1",
  expr,
  height = 400,
}) => {
  const [startTime, setStartTime] = React.useState<number>(
    now.subtract(1, "hour").unix()
  );
  const [endTime, setEndTime] = React.useState<number>(now.unix());
  const [resolution, setResolution] = React.useState<number>(14);
  const [data, setData] = React.useState<PromValue[]>([]);
  const [tabKey, setTabKey] = React.useState<string>("table");
  const [loading, setLoading] = React.useState<boolean>(false);
  const [dateType, setDateType] = React.useState<DateDataType>("date");
  const [err, setErr] = React.useState<string | undefined>();
  const [datasource, setDatasource] = React.useState<promData[]>([]);

  const handleTabChange = (key: string) => {
    setTabKey(key);
    switch (key) {
      case "graph":
        setEndTime(now.unix());
        setStartTime(now.subtract(1, "hour").unix());
        setDateType("range");
        break;
      case "table":
        setEndTime(now.unix());
        setDateType("date");
        break;
    }
  };

  const fetchValues = async (tabKey: string) => {
    if (!expr) return;
    let path = "";
    const abortController = new AbortController();
    const params: URLSearchParams = new URLSearchParams({
      query: expr,
    });
    switch (tabKey) {
      case "graph":
        path = "query_range";
        params.append("start", startTime.toString());
        params.append("end", endTime.toString());
        params.append("step", resolution.toString());
        break;
      case "table":
        path = "query";
        params.append("time", endTime.toString());
        break;
      default:
        throw new Error('Invalid panel type "' + tabKey + '"');
    }

    setLoading(true);
    let query = await fetch(`${pathPrefix}/${apiPath}/${path}?${params}`, {
      cache: "no-store",
      credentials: "same-origin",
      signal: abortController.signal,
    })
      .then((resp) => resp.json())
      .catch((err) => {
        console.log("err", err);
      })
      .finally(() => setLoading(false));

    if (query.status !== "success") {
      setErr(query.error);
      return;
    }

    setDatasource(query.data.result);
    setErr(undefined);

    if (query.data) {
      const { result } = query.data;
      setData([...result]);
    }
  };

  const handleSearch = (
    type: DateDataType,
    values: number | [number, number],
    step?: number
  ) => {
    switch (type) {
      case "date":
        setEndTime(values as number);
        break;
      case "range":
        setStartTime((values as [number, number])[0]);
        setEndTime((values as [number, number])[1]);
        setResolution(step || 14);
        break;
    }
  };

  useEffect(() => {
    if (!visible || !expr) return;
    fetchValues(tabKey);
  }, [expr, visible, tabKey, startTime, endTime, resolution]);

  const getValues = (val: PromValue): any => {
    if (val.value && !Array.isArray(val.value[1])) {
      return val.value[1];
    }

    if (
      val.values &&
      val.values.length > 0 &&
      Array.isArray(val.values[0]) &&
      val.values[0].length === 2 &&
      !Array.isArray(val.values[0][1])
    ) {
      return val.values[0][1];
    }

    return "";
  };

  const renderGraph = (metricValues?: promData[]) => {
    if (!metricValues) return null;
    return (
      <AreaStackGradient
        height={height}
        data={metricValues}
        id="prom-metric-data-chart"
        title={expr}
      />
    );
  };

  return (
    <Modal
      visible={visible}
      onCancel={onCancel}
      style={{ width: "80%" }}
      footer={null}
      unmountOnExit
    >
      <SearchForm type={dateType} onSearch={handleSearch} />
      <Tabs
        direction="horizontal"
        onChange={handleTabChange}
        defaultActiveTab="table"
      >
        <TabPane
          title={
            <Button type="text">
              <IconList />
              指标列表
            </Button>
          }
          key="table"
        >
          {err && <Alert closable type="error" content={err} />}
          <List
            loading={loading}
            style={{ height: height }}
            dataSource={data}
            render={(item, index) => {
              return (
                <List.Item key={index} id={`list-${index}`}>
                  <Space direction="horizontal" style={{ width: "100%" }}>
                    <CopyText
                      showMessage
                      mode="button"
                      message="Copied"
                      style={{
                        width: "100%",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {item?.metric
                        ? `${item?.metric?.__name__ || ""}{${Object.keys(
                            item.metric
                          )
                            .filter((key) => key !== "__name__" && key !== "id")
                            .map((key) => `${key}="${item.metric[key]}"`)
                            .join(", ")}}`
                        : expr}
                    </CopyText>
                    <div style={{ float: "right" }}>{getValues(item)}</div>
                  </Space>
                </List.Item>
              );
            }}
          />
        </TabPane>
        <TabPane
          title={
            <Button type="text">
              <IconImage />
              指标图表
            </Button>
          }
          key="graph"
        >
          <div
            style={{
              height: height,
              overflowY: "auto",
              overflowX: "hidden",
            }}
          >
            {err ? (
              <Alert
                closable
                type="error"
                // title="Error"
                content={err}
              />
            ) : (
              renderGraph([...datasource])
            )}
            {/*<h1>该功能完成进度</h1>*/}
            {/*<Progress percent={20} size={"large"} />*/}
          </div>
        </TabPane>
      </Tabs>
    </Modal>
  );
};

export default PromValueModal;
