import React, { useEffect, useState } from "react";
import { GroupItem } from "@/apis/prom/prom";
import {
  defaultListStrategyRequest,
  ListStrategyRequest,
} from "@/apis/prom/strategy/strategy";
import ShowTable from "@/pages/strategy/child/ShowTable";
import { Button, Divider, Form, Grid, Input } from "@arco-design/web-react";
import StrategyModal, {
  StrategyModalProps,
  StrategyValues,
} from "@/pages/strategy/child/StrategyModal";
import { StrategyCreate } from "@/apis/prom/strategy/api";

import styles from "./style/strategy.module.less";

export interface StrategyListProps {
  groupItem?: GroupItem;
}

const pathPrefix = "http://localhost:9090";

const { Row, Col } = Grid;

const Strategy: React.FC<StrategyListProps> = (props) => {
  const { groupItem } = props;

  const [strategyModalProps, setStrategyModalProps] = useState<
    StrategyModalProps | undefined
  >();
  const [strategyModalVisabled, setStrategyModalVisabled] =
    useState<boolean>(false);
  const [refresh, setRefresh] = useState<boolean>(false);
  const [queryParams, setQueryParams] = React.useState<ListStrategyRequest>(
    groupItem
      ? {
          ...defaultListStrategyRequest,
          strategy: {
            groupId: groupItem.id,
          },
        }
      : defaultListStrategyRequest
  );

  const openAddModalForm = () => {
    setStrategyModalVisabled(true);
    setStrategyModalProps({
      title: "添加报警规则",
      setVisible: setStrategyModalVisabled,
      initialValues: {
        datasource: pathPrefix,
        for: "60m",
        groupId: groupItem?.id,
      },
      onOk: async (strategyValues: StrategyValues) => {
        const resp = await StrategyCreate({
          alarmPageIds: strategyValues.alarmPageIds || [],
          alert: strategyValues.alert || "",
          alertLevelId: strategyValues.alertLevelId || 0,
          annotations: strategyValues.annotations || {},
          categorieIds: strategyValues.categorieIds || [],
          expr: strategyValues.expr || "",
          for: strategyValues.for || "60m",
          groupId: strategyValues.groupId || 0,
          labels: strategyValues.labels || {},
        });
        setRefresh?.((prevState) => !prevState);
        return resp;
      },
    });
  };

  useEffect(() => {
    if (!groupItem) return;
    setQueryParams({
      ...queryParams,
      strategy: {
        groupId: groupItem.id,
      },
    });
  }, [groupItem]);

  return (
    <div className={styles.strategyDiv}>
      <Form layout="inline" className={styles.queryForm}>
        <Form.Item label="规则名称" field="alert">
          <Input placeholder="通过规则名称模糊搜索" />
        </Form.Item>
      </Form>
      <Row className={styles.optionDiv}>
        <Col span={12}>
          <Button type="primary" onClick={openAddModalForm}>
            添加规则
          </Button>
        </Col>
        <Col span={12}></Col>
      </Row>
      <ShowTable
        database={pathPrefix}
        setQueryParams={setQueryParams}
        queryParams={queryParams}
        setRefresh={setRefresh}
        setStrategyModalProps={setStrategyModalProps}
        setStrategyModalVisabled={setStrategyModalVisabled}
        refresh={refresh}
      />
      <StrategyModal {...strategyModalProps} visible={strategyModalVisabled} />
    </div>
  );
};

export default Strategy;
