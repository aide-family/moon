import React, { useEffect } from "react";
import { randomString } from "@/utils/random";
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

  const [__id__] = React.useState<string>(`area-stack-gradient-${id}`);

  useEffect(() => {
    const chartDom = document.getElementById(__id__);
    if (!chartDom) return;
    const myChart = echarts.init(chartDom);
    const options = buildOption(data, title);
    options && myChart.setOption(options);
  }, [__id__, data]);

  return <div id={__id__} style={{ width: "100%", height: height }}></div>;
};

export default AreaStackGradient;
