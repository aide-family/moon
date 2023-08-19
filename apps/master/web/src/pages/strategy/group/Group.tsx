import React, { useEffect } from "react";

import SearchForm, {
  SearchFormType,
} from "@/pages/strategy/group/child/SearchForm";
import type { sizeType } from "@/pages/strategy/group/child/ShowTable";
import ShowTable from "@/pages/strategy/group/child/ShowTable";
import OptionLine from "@/pages/strategy/group/child/OptionLine";
import {
  defaultListGroupRequest,
  ListGroupRequest,
} from "@/apis/prom/group/group";
import { useSearchParams } from "react-router-dom";

import groupStyle from "./style/group.module.less";

const Group: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [queryParams, setQueryParams] = React.useState<
    ListGroupRequest | undefined
  >();
  const [tableSize, setTableSize] = React.useState<sizeType>("default");
  const [refresh, setRefresh] = React.useState<boolean>(false);

  const handleSearchChange = (params: SearchFormType) => {
    setQueryParams((prev?: ListGroupRequest): ListGroupRequest => {
      return {
        ...prev,
        query: {
          ...prev?.query,
          page: defaultListGroupRequest.query?.page,
          endAt: params.endAt ? params.endAt + "" : undefined,
          startAt: params.startAt ? params.startAt + "" : undefined,
          keyword: params.keyword,
        },
        group: {
          ...prev?.group,
          strategyCount: params.strategyCount,
          status: params.status,
        },
      };
    });
  };

  const handleOnRefresh = () => {
    setRefresh((prev) => !prev);
  };

  useEffect(() => {
    try {
      setQueryParams(() => {
        let q = searchParams.get("q");
        if (!q) return defaultListGroupRequest;
        return JSON.parse(q || "");
      });
    } catch (e) {}
  }, []);

  return (
    <div className={groupStyle.GroupDiv}>
      <SearchForm onChange={handleSearchChange} />
      <OptionLine onTableSizeChange={setTableSize} refresh={handleOnRefresh} />
      <ShowTable
        size={tableSize}
        refresh={refresh}
        queryParams={queryParams}
        setQueryParams={setQueryParams}
      />
    </div>
  );
};

export default Group;
