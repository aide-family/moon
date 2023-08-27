import React, { useState } from "react";
import {
  IconArrowRise,
  IconTrophy,
  IconUser,
} from "@arco-design/web-react/icon";
import {
  Button,
  Divider,
  Drawer,
  Space,
  Avatar,
  Statistic,
} from "@arco-design/web-react";

import styles from "../style/alarm.module.less";
import AreaStackGradient from "@/components/charts/area-stack-gradient";
import { data } from "@/components/charts/area-stack-gradient/data";
import StatisticsChart from "@/pages/home/child/StatisticsChart";

export interface StatisticsDrawerProps {
  visible?: boolean;
  setVisible?: React.Dispatch<React.SetStateAction<boolean>>;
}

export interface NodeProps {
  title?: React.ReactNode;
  value?: number | string;
  color?: string;
}

const Node: React.FC<NodeProps> = (props) => {
  const { title, value } = props;
  return (
    <div className={styles.node}>
      <Avatar style={{ backgroundColor: "#3370ff" }}>
        <IconUser />
      </Avatar>

      <Statistic
        className={styles.statistic}
        title={<div style={{ marginBottom: -8 }}>{title}</div>}
        value={value}
        suffix={<IconArrowRise style={{ color: "#ee4d38" }} />}
      />
    </div>
  );
};

const StatisticsDrawer: React.FC<StatisticsDrawerProps> = (props) => {
  const [visible, setVisible] = useState(false);

  const openDrawer = () => {
    setVisible(true);
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const nodes: NodeProps[] = [
    {
      title: "报警总数",
      value: 45342,
    },
    {
      title: "最近报警总数",
      value: 231,
    },
    {
      title: "最近恢复总数",
      value: 531,
    },
    {
      title: "报警规则TOP1",
      value: "prometheus-alert-1",
    },
    {
      title: "报警规则TOP2",
      value: "prometheus-alert-2",
    },
    {
      title: "报警规则TOP3",
      value: "prometheus-alert-3",
    },
  ];

  return (
    <>
      <div>
        <Button type="default" icon={<IconTrophy />} onClick={openDrawer} />
        <Drawer
          className={styles.alarmStatisticsDrawer}
          escToExit
          focusLock
          title={null}
          height={600}
          visible={visible}
          placement="top"
          onCancel={handleCancel}
          footer={null}
          closable={false}
        >
          <div className={styles.aggregate}>
            {nodes.map((node, index) => {
              return <Node key={index} {...node} />;
            })}
          </div>
          <div>
            <StatisticsChart />
          </div>
        </Drawer>
      </div>
    </>
  );
};

export default StatisticsDrawer;
