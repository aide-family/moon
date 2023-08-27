import React, { useEffect, useState } from "react";
import { GroupItem, PromStrategyItemRequest } from "@/apis/prom/prom";
import {
  defaultListStrategyRequest,
  ListStrategyRequest,
} from "@/apis/prom/strategy/strategy";
import ShowTable from "@/pages/strategy/child/ShowTable";
import { Button, Form, Grid } from "@arco-design/web-react";
import StrategyModal, {
  StrategyModalProps,
  StrategyValues,
} from "@/pages/strategy/child/StrategyModal";
import { StrategyCreate } from "@/apis/prom/strategy/api";

import styles from "./style/strategy.module.less";
import QueryForm from "@/pages/strategy/child/QueryForm";

export interface StrategyListProps {
  groupItem?: GroupItem;
}

const pathPrefix = "http://localhost:9090";

const { Row, Col } = Grid;

const Strategy: React.FC<StrategyListProps> = (props) => {
  const { groupItem } = props;
  const [searchForm] = Form.useForm();

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

  const setSearchParams = (value: PromStrategyItemRequest) => {
    setQueryParams({
      ...queryParams,
      query: {
        ...queryParams.query,
        keyword: value.alert,
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
      <div className={styles.queryForm}>
        <QueryForm form={searchForm} onChange={setSearchParams} />
      </div>

      <Row className={styles.optionDiv}>
        <Col span={12}>
          <Button type="primary" onClick={openAddModalForm}>
            添加规则
          </Button>
        </Col>
        <Col span={12}></Col>
      </Row>
      <ShowTable
        className={styles.strategyTable}
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
