import React, { useEffect, useState } from "react";
import { GroupItem, PromStrategyItemRequest } from "@/apis/prom/prom";
import {
  defaultListStrategyRequest,
  ListStrategyRequest,
} from "@/apis/prom/strategy/strategy";
import type { tableSize } from "@/pages/strategy/child/ShowTable";
import ShowTable from "@/pages/strategy/child/ShowTable";
import { Button, Form, Grid, Radio } from "@arco-design/web-react";
import StrategyModal, {
  StrategyModalProps,
  StrategyValues,
} from "@/pages/strategy/child/StrategyModal";
import { StrategyCreate } from "@/apis/prom/strategy/api";

import styles from "./style/strategy.module.less";
import QueryForm from "@/pages/strategy/child/QueryForm";
import { sizes } from "@/pages/strategy/group/child/OptionLine";
import groupStyle from "@/pages/strategy/group/style/group.module.less";

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
  const [tableSize, setTableSize] = useState<tableSize>("small");
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

  const onTableSizeChange = (size: tableSize) => {
    setTableSize(size);
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

      <div className={groupStyle.OptionLineDiv}>
        <Row className={groupStyle.Row}>
          <Col span={12} className={groupStyle.LeftCol}>
            <Button type="primary" onClick={openAddModalForm}>
              添加规则
            </Button>
          </Col>
          <Col span={12} className={groupStyle.RightCol}>
            <Radio.Group
              type="button"
              options={sizes}
              defaultValue={(sizes.length && sizes[0].value) || "default"}
              onChange={onTableSizeChange}
            />
          </Col>
        </Row>
      </div>

      <ShowTable
        size={tableSize}
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
