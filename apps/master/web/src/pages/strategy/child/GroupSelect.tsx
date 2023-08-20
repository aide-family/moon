import React, { useCallback, useEffect, useRef, useState } from "react";
import { SimpleItem } from "@/apis/prom/prom";
import { Select, Spin } from "@arco-design/web-react";
import debounce from "lodash/debounce";
import { GroupSimpleList } from "@/apis/prom/group/api";

export interface GroupSelectProps {
  value?: number;
  onChange?: (val: number) => void;
  defaultValue?: number;
  disabled?: boolean;
}

type Option = {
  label: React.ReactNode;
  value: number | string;
};

const GroupSelect: React.FC<GroupSelectProps> = (props) => {
  const { value, onChange, disabled, defaultValue } = props;

  const [options, setOptions] = useState<Option[]>([]);
  const [fetching, setFetching] = useState(false);

  const getGroupSimpleList = (keyword?: string) => {
    setFetching(true);
    GroupSimpleList({
      page: { current: 1, size: 10 },
      keyword: keyword,
    })
      .then((resp) => {
        const { groups } = resp;
        setOptions(
          (groups || []).map((item) => ({
            label: item.name,
            value: item.id,
          }))
        );
      })
      .finally(() => setFetching(false));
  };

  const debouncedFetchUser = useCallback(
    debounce((inputValue: string) => {
      getGroupSimpleList(inputValue);
    }, 500),
    []
  );

  useEffect(() => {
    getGroupSimpleList();
  }, []);

  return (
    <>
      <Select
        placeholder="请选择规则组"
        showSearch
        allowClear
        disabled={disabled}
        value={value}
        filterOption={false}
        defaultValue={defaultValue}
        onChange={onChange}
        options={options}
        notFoundContent={
          fetching ? (
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <Spin style={{ margin: 12 }} />
            </div>
          ) : null
        }
        onSearch={debouncedFetchUser}
      />
    </>
  );
};

export default GroupSelect;
