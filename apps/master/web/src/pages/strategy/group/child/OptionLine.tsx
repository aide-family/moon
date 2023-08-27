import React, { useEffect } from "react";
import AddGroup from "@/pages/strategy/group/child/AddGroup";
import groupStyle from "@/pages/strategy/group/style/group.module.less";
import { Button, Dropdown, Grid, Menu, Radio } from "@arco-design/web-react";
import { IconDown, IconRefresh } from "@arco-design/web-react/icon";
import type { sizeType } from "@/pages/strategy/group/child/ShowTable";

export interface OptionLineProps {
  // 刷新
  refresh?: () => void;
  onTableSizeChange?: (size: sizeType) => void;
}

const { Row, Col } = Grid;

// 'default', 'middle', 'small', 'mini'
export const sizes = [
  {
    label: "默认",
    value: "default",
  },
  {
    label: "中等",
    value: "middle",
  },
  {
    label: "小",
    value: "small",
  },
  {
    label: "迷你",
    value: "mini",
  },
];

const defaultInterval = 60;

const OptionLine: React.FC<OptionLineProps> = (props) => {
  const { refresh, onTableSizeChange } = props;

  const [refreshInterval, setRefreshInterval] =
    React.useState<number>(defaultInterval);

  const handleIntervalChange = (key: string) => {
    setRefreshInterval(+key);
  };

  const dropList = (
    <Menu onClickMenuItem={handleIntervalChange}>
      <Menu.Item key="10">10s</Menu.Item>
      <Menu.Item key="30">30s</Menu.Item>
      <Menu.Item key="60">60s</Menu.Item>
    </Menu>
  );

  useEffect(() => {
    const timer = setInterval(() => {
      refresh && refresh();
    }, refreshInterval * 1000);
    return () => clearInterval(timer);
  }, [refresh, refreshInterval]);

  return (
    <>
      <div className={groupStyle.OptionLineDiv}>
        <Row className={groupStyle.Row}>
          <Col span={12} className={groupStyle.LeftCol}>
            <AddGroup onFinished={refresh}>
              <Button type="primary">添加分组</Button>
            </AddGroup>
          </Col>
          <Col span={12} className={groupStyle.RightCol}>
            <Button.Group>
              <Button type="default" onClick={refresh} icon={<IconRefresh />} />
              <Dropdown droplist={dropList} position="br">
                <Button type="default">
                  {refreshInterval}s<IconDown />
                </Button>
              </Dropdown>
            </Button.Group>

            <Radio.Group
              type="button"
              options={sizes}
              defaultValue={(sizes.length && sizes[0].value) || "default"}
              onChange={onTableSizeChange}
            />
          </Col>
        </Row>
      </div>
    </>
  );
};

export default OptionLine;
