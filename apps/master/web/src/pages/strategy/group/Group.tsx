import React, {useEffect} from "react";

import SearchForm, {SearchFormType} from "@/pages/strategy/group/child/SearchForm";
import ShowTable from "@/pages/strategy/group/child/ShowTable";
import OptionLine from "@/pages/strategy/group/child/OptionLine";
import {defaultListGroupRequest, ListGroupRequest} from "@/apis/prom/group/group";
import {useSearchParams} from "react-router-dom";

import groupStyle from "./style/group.module.less";

const Group: React.FC = () => {
    const [searchParams] = useSearchParams();

    const [queryParams, setQueryParams] = React.useState<ListGroupRequest | undefined>();
    const [isNoData, setIsNoData] = React.useState<boolean>(false);

    const handleSearchChange = (params: SearchFormType) => {
        setIsNoData(false);
        setQueryParams((prev?: ListGroupRequest): ListGroupRequest => {
            return {
                ...prev,
                query: {
                    ...prev?.query,
                    page: prev?.query.page || defaultListGroupRequest.query?.page,
                    endAt: params.endAt ? params.endAt + "" : undefined,
                    startAt: params.startAt ? params.startAt + "" : undefined,
                    keyword: params.keyword,
                },
                group: {
                    ...prev?.group,
                    strategyCount: params.strategyCount,
                    status: params.status,
                },
            }
        });
    }

    useEffect(() => {
        try {
            setQueryParams(() => {
                let q = searchParams.get("q")
                if (!q) return defaultListGroupRequest
                return JSON.parse(q || "")
            })
        } catch (e) {
        }
    }, [])

    return (
        <div className={groupStyle.GroupDiv}>
            <SearchForm onChange={handleSearchChange}/>
            <OptionLine/>
            <ShowTable
                queryParams={queryParams}
                setQueryParams={setQueryParams}
                setIsNoData={setIsNoData}
                isNoData={isNoData}
            />
        </div>
    );
};

export default Group;