import React from "react";
import { Button, Grid, Radio } from "@arco-design/web-react";

import styles from "../style/alarm.module.less";
import { sizes } from "@/pages/strategy/group/child/OptionLine";
import { AlarmModalProps } from "@/pages/home/child/AlarmModal";
import { Response } from "@/apis/type";
import { Post } from "@/apis/requests";
import AlarmPageTab from "@/pages/home/child/AlarmPageTab";
import { IconRefresh, IconTrophy } from "@arco-design/web-react/icon";
import StatisticsDrawer from "@/pages/home/child/StatisticsDrawer";

export type sizeType = "default" | "middle" | "small" | "mini";

export interface AlarmOptionProps {
  setSize?: React.Dispatch<React.SetStateAction<sizeType>>;
  setAlarmModalProps?: React.Dispatch<
    React.SetStateAction<AlarmModalProps | undefined>
  >;
  setVisible?: React.Dispatch<React.SetStateAction<boolean>>;
  refresh?: () => void;
}

const { Row, Col } = Grid;

const AlarmOption: React.FC<AlarmOptionProps> = (props) => {
  const { setSize, refresh, setAlarmModalProps, setVisible } = props;

  const onTableSizeChange = (size: sizeType) => {
    setSize?.(size);
  };

  const openAlarmModal = () => {
    setVisible?.(true);
    setAlarmModalProps?.({
      onOk: async (values: any) => {
        // TODO
        const resp = await Post<Response>("", values);
        refresh?.();
        return resp;
      },
    });
  };

  return (
    <>
      <div className={styles.alarmOption}>
        <Row className={styles.Row}>
          <Col span={6} className={styles.LeftCol}>
            <Button type="primary" onClick={openAlarmModal}>
              添加报警页面
            </Button>
          </Col>
          <Col span={12}>
            <AlarmPageTab />
          </Col>
          <Col span={6} className={styles.RightCol}>
            <StatisticsDrawer />
            <Button type="default" icon={<IconRefresh />} />
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

export default AlarmOption;
