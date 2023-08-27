import React, { useEffect, useState } from "react";
import { randomString } from "@/utils/random";
import type { EChartsType } from "echarts";
import * as echarts from "echarts";
import { buildOption } from "@/components/charts/area-stack-gradient/option";

export interface AreaStackGradientProps {
  height?: number | string;
  data: any;
  id?: string;
  title?: string;
}

const AreaStackGradient: React.FC<AreaStackGradientProps> = (props) => {
  const { height = 600, data, id = randomString(10), title } = props;

  const [__id__] = useState<string>(`area-stack-gradient-${id}`);
  const [myChart, setMyChart] = useState<EChartsType>();

  useEffect(() => {
    const chartDom = document.getElementById(__id__);
    if (!chartDom) return;
    setMyChart(echarts.init(chartDom));
  }, []);

  useEffect(() => {
    if (myChart) {
      const options = buildOption(data, title);
      options && myChart?.setOption(options);
    }
  }, [__id__, data]);

  return <div id={__id__} style={{ width: "100%", height: height }}></div>;
};

export default AreaStackGradient;
