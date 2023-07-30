import * as echarts from "echarts";
import dayjs from "dayjs";

export type promData = {
  metric?: {
    __name__: string;
    [key: string]: string;
  };
  values?: [number, string][];
};

export const buildOption = (
  data?: promData[],
  title?: string
): echarts.EChartsOption | undefined => {
  if (!data || !data?.length) return undefined;
  const lineNames = data.map((m) => {
    return m?.metric
      ? `${m?.metric?.__name__ || ""}{${Object.keys(m.metric)
          ?.filter((n) => n !== "__name__")
          ?.map((n) => `${n}="${m.metric?.[n]}"`)
          ?.join(" ,")}}`
      : "";
  });

  const lineData = data
    .find((_: any, index: number) => {
      return index === 0;
    })
    ?.values?.map((m: (string | number)[]) => {
      return dayjs(+m[0] * 1000).format("YYYY-MM-DD HH:mm:ss");
    });

  const dataList: any[] = data.map((m, index) => {
    return {
      data: m?.values?.map((n) => {
        return +n[1];
      }),
      name: lineNames[index],
      type: "line",
      smooth: true, // 是否平滑曲线显示
    };
  });

  return {
    // 显示网格

    grid: {
      // left: "3%",
      // right: "4%",
      // bottom: "25%",
      // top: "30%",
      // containLabel: true,
    },

    tooltip: {
      trigger: "item",
      // alwaysShowContent: true,
    },
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: "none",
          title: {
            zoom: "区域缩放",
            back: "区域缩放还原",
          },
        },

        magicType: {
          type: ["line", "bar", "stack"],
          title: {
            line: "折线图",
            bar: "柱状图",
            stack: "堆叠",
          },
        },
        saveAsImage: {
          title: "保存为图片",
        },
      },
    },
    xAxis: {
      type: "category",
      data: lineData,
      // 显示坐标线
      axisLine: {
        show: true,
      },
    },
    yAxis: {
      type: "value",
      // 显示坐标线
      axisLine: {
        show: true,
      },
    },
    legend: {
      data: lineNames,
      // 支持滑动
      type: "scroll",
      // 位置
      orient: "horizontal",
      // right: 0,
      // top: 0,
      bottom: 0,
      // width: 100%,
      height: 400,
      textStyle: {
        overflow: "break",
      },
    },
    series: dataList,
  };
};
