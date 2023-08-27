import React, { ReactNode } from "react";
import { AlarmQueryParams } from "@/apis/prom/alarm/alarm";
import { Radio } from "@arco-design/web-react";

import styles from "../style/alarm.module.less";

export interface AlarmPageTabProps {
  setQueryParams?: React.Dispatch<React.SetStateAction<AlarmQueryParams>>;
}

export type radioGroupOption =
  | string
  | number
  | {
      label: ReactNode;
      value: any;
      disabled?: boolean;
    };

const AlarmPage: React.FC<AlarmPageTabProps> = (props) => {
  const pageOptions: radioGroupOption[] = ["系统告警", "PCDN业务", "APP监控"];

  return (
    <>
      <div className={styles.alarmPageTab}>
        <Radio.Group type="button" options={pageOptions} />
      </div>
    </>
  );
};

export default AlarmPage;
