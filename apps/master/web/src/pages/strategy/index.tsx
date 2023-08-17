import React, {useEffect} from "react";
import "./style/index.less";
import {GroupItem} from "@/apis/prom/prom";
import {defaultListStrategyRequest, ListStrategyRequest} from "@/apis/prom/strategy/strategy";
import ShowTable from "@/pages/strategy/child/ShowTable";

export type DataSourceType = {
    id: number;
    name: string;
    group: string;
    expr: string;
    labels: { [key: string]: string };
    annotations: { [key: string]: string };
    for: string;
    datasource: string;
    // 优先级
    priority?: number;
};

export interface StrategyListProps {
    groupItem?: GroupItem
}

const pathPrefix = "http://localhost:9090";

const Strategy: React.FC<StrategyListProps> = (props) => {
    const {groupItem} = props;

    const [queryParams, setQueryParams] = React.useState<ListStrategyRequest>(groupItem ? {
        ...defaultListStrategyRequest,
        strategy: {
            groupId: groupItem.id
        }
    } : defaultListStrategyRequest);

    useEffect(() => {
        if (!groupItem) return;
        setQueryParams({
            ...queryParams,
            strategy: {
                groupId: groupItem.id
            }
        })
    }, [groupItem]);

    return (
        <div>
            <ShowTable database={pathPrefix} setQueryParams={setQueryParams} queryParams={queryParams}/>
        </div>
    );
};

export default Strategy;
