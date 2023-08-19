import React from "react";
import type { GroupItem } from "@/apis/prom/prom";
import Strategy from "@/pages/strategy";

export interface StrategyListProps {
  groupItem?: GroupItem;
}

const StrategyList: React.FC<StrategyListProps> = (props) => {
  const { groupItem } = props;

  return <Strategy groupItem={groupItem} />;
};

export default StrategyList;
