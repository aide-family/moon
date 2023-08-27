import React, { useEffect, useState } from "react";
import { randomString } from "@/utils/random";
import * as echarts from "echarts";
import { EChartsType } from "echarts";
import dayjs from "dayjs";

export interface StatisticsChartProps {
  height?: number | string;
  id?: string;
  title?: string;
}

let _data = [
  {
    name: dayjs().add(-5, "hour").format("YYYY-MM-DD HH:mm:ss"),
    value1: 2553,
    value2: 3133,
  },
  {
    name: dayjs().add(-4, "hour").format("YYYY-MM-DD HH:mm:ss"),
    value1: 3233,
    value2: 2133,
  },
  {
    name: dayjs().add(-3, "hour").format("YYYY-MM-DD HH:mm:ss"),
    value1: 4200,
    value2: 3100,
  },
  {
    name: dayjs().add(-2, "hour").format("YYYY-MM-DD HH:mm:ss"),
    value1: 4180,
    value2: 4280,
  },
  {
    name: dayjs().add(-1, "hour").format("YYYY-MM-DD HH:mm:ss"),
    value1: 2199,
    value2: 3299,
  },
  {
    name: dayjs().format("YYYY-MM-DD HH:mm:ss"),
    value1: 2253,
    value2: 1133,
  },
];

let _dataTitle = [
  {
    name: "报警个数",
  },
  {
    name: "恢复个数",
  },
];

let xLabel = _data.map((x) => x.name); //['7/13','7/14', '7/15', '7/16', '7/17']
let res1 = _data.map((x) => x.value1); // [22553, 3233, 4200, 4180, 2199];
let res2 = _data.map((x) => x.value2); //[3133, 2133, 3100, 4280, 3299];

const StatisticsChart: React.FC<StatisticsChartProps> = (props) => {
  const { height = 400, id = randomString(10), title } = props;

  const [__id__] = useState<string>(`statictics-chart-${id}`);
  const [myChart, setMyChart] = useState<EChartsType>();

  useEffect(() => {
    const chartDom = document.getElementById(__id__);
    if (!chartDom) return;
    setMyChart(echarts.init(chartDom));
  }, []);

  useEffect(() => {
    if (myChart) {
      const options = {
        backgroundColor: "#fff",
        title: {
          show: true,
          text: "报警统计",
          left: "50%",
          top: "5%",
          color: "rgba(126,199,255,1)",
          fontStyle: "oblique",
          fontSize: 14,
        },
        tooltip: {
          trigger: "axis",
          backgroundColor: "transparent",
          axisPointer: {
            lineStyle: {
              color: {
                type: "linear",
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  {
                    offset: 0,
                    color: "rgba(126,199,255,0)", // 0% 处的颜色
                  },
                  {
                    offset: 0.5,
                    color: "rgba(126,199,255,1)", // 100% 处的颜色
                  },
                  {
                    offset: 1,
                    color: "rgba(126,199,255,0)", // 100% 处的颜色
                  },
                ],
                global: false, // 缺省为 false
              },
            },
          },
          formatter: (p: any) => {
            return `<div >
        <div style="padding: 4px 8px 4px 14px;display: flex;
        justify-content: center;
        align-items: center;
        flex-direction: column;position: relative;z-index: 1;">
            <div style="margin-bottom: 4px;width:100%;display:${
              p[0] ? "flex" : "none"
            };justify-content:space-between;align-items:center;color:rgb(255,128,0);">
                <span style="font-size:14px;color:rgb(255,128,0)">${
                  p[0] ? p[0].seriesName : ""
                }</span>：
                <span style="font-size:14px;color:rgb(255,128,0);">${
                  p[0] ? p[0].data : ""
                }</span>
            </div>
            <div style="width:100%;height:100%;display:${
              p[1] ? "flex" : "none"
            };justify-content:space-between;align-items:center;color:#7ec7ff;padding-left:4px;">
                <span style="font-size:14px;color:#7ec7ff;">${
                  p[1] ? p[1].seriesName : ""
                }</span>：
                <span style="font-size:14px;color:#7ec7ff;">${
                  p[1] ? p[1].data : ""
                }</span>
            </div>
        </div>
    </div>`;
          },
        },
        legend: {
          align: "left",
          right: "10%",
          top: "5%",
          type: "plain",
          color: "#000",
          fontSize: 16,
          icon: "rect",
          itemGap: 25,
          itemWidth: 18,

          data: _dataTitle,
        },

        grid: {
          top: "15%",
          left: "10%",
          right: "10%",
          bottom: "15%",
          // containLabel: true
        },
        xAxis: [
          {
            type: "category",
            boundaryGap: true,
            axisLine: {
              //坐标轴轴线相关设置。数学上的x轴
              show: true,
              lineStyle: {
                color: "#rgb(77,83,141)",
              },
            },
            axisLabel: {
              //坐标轴刻度标签的相关设置
              color: "#7ec7ff",
              padding: 16,
              fontSize: 14,
              formatter: function (data: any) {
                return data;
              },
            },
            axisTick: {
              show: false,
            },
            data: xLabel,
          },
        ],
        yAxis: [
          {
            nameTextStyle: {
              color: "#7ec7ff",
              fontSize: 16,
              padding: 10,
            },
            axisLine: {
              show: true,
              lineStyle: {
                color: "#rgb(77,83,141)",
              },
            },
            axisLabel: {
              show: true,
              color: "#7ec7ff",
              padding: 8,
            },
            axisTick: {
              show: false,
            },
          },
        ],

        series: [
          {
            name: _dataTitle[0].name,
            type: "line",
            symbol: "triangle", // 默认是空心圆（中间是白色的），改成实心圆
            showAllSymbol: true,
            showSymbol: true,
            smooth: true, //平滑
            symbolSize: 10,
            label: {
              show: true,
              position: "top",
              color: "rgb(255,128,0)",
            },
            lineStyle: {
              width: 1,
              color: "rgb(255,128,0)", // 线条颜色
            },
            itemStyle: {
              color: "rgb(255,128,0)",
            },
            tooltip: {
              show: true,
            },

            data: res1,
          },
          {
            name: _dataTitle[1].name,
            type: "line",
            symbol: "rect", // 默认是空心圆（中间是白色的），改成实心圆
            showAllSymbol: true,
            showSymbol: true,
            smooth: true, //平滑
            symbolSize: 10,
            label: {
              show: true,
              position: "bottom",
              color: "rgba(25,163,223,1)",
            },
            lineStyle: {
              width: 1,
              color: "rgba(25,163,223,1)", // 线条颜色
              type: "solid",
            },
            itemStyle: {
              color: "rgba(25,163,223,1)",
            },
            tooltip: {
              show: true,
            },

            data: res2,
          },
        ],
      };
      options && myChart?.setOption(options);
    }
  }, [__id__, myChart]);

  return (
    <>
      <div id={__id__} style={{ width: "100%", height: height }}></div>
    </>
  );
};

export default StatisticsChart;
