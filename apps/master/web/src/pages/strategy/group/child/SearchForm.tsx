import React, {useEffect} from "react";
import {DatePicker, Form, Input, Radio, Select} from "@arco-design/web-react";
import {Status, StatusMap} from "@/apis/prom/prom";
import dayjs from "dayjs";
import {useSearchParams} from "react-router-dom";
import groupStyle from "@/pages/strategy/group/style/group.module.less";

export type SearchFormType = {
    keyword?: string;
    status?: number;
    strategyCount?: number;
    startAt?: number;
    endAt?: number;
}

export interface SearchFormProps {
    onChange?: (value: SearchFormType) => void;
}

const strategyCounts: number[] = [10, 20, 30, 40, 50, 60, 70, 80, 90];
const defaultWidth = 300;

const SearchForm: React.FC<SearchFormProps> = (props) => {
    const {
        onChange,
    } = props;
    const [form] = Form.useForm();
    const [searchParams] = useSearchParams();

    const handleFormChange = (values: any) => {
        // 判断time是否存在, 如果存在, 则转换成时间戳赋值到values
        if (values?.time && values?.time?.length === 2) {
            const [start, end] = values.time;
            values.startAt = dayjs(start).unix();
            values.endAt = dayjs(end).unix();
            delete values.time;
        }
        if (!values?.status) {
            delete values.status;
        }
        if (!values?.keyword) {
            delete values.keyword;
        }
        if (!values?.strategyCount) {
            delete values.strategyCount;
        }

        onChange?.(values);
    }

    useEffect(() => {
        let q = searchParams.get("q")
        if (!q) return;
        try {
            let query = JSON.parse(q || "{}")
            let formData = {
                keyword: query?.query.keyword,
                strategyCount: query?.group?.strategyCount,
                time: [
                    query?.query.startAt ? dayjs.unix(+query.query.startAt).format("YYYY-MM-DD HH:mm:ss") : "",
                    query?.query.endAt ? dayjs.unix(+query.query.endAt).format("YYYY-MM-DD HH:mm:ss") : "",
                ],
                status: query?.group?.status || 0,
            }
            form.setFieldsValue(formData)
        } catch (e) {
        }
    }, [])

    return <div className={groupStyle.SearchFormDiv}>
        <Form
            form={form}
            layout="inline"
            onValuesChange={(_, values) => handleFormChange(values)}
        >
            <Form.Item field="keyword" label="名称">
                <Input placeholder="根据名称模糊查询" style={{width: defaultWidth}} allowClear autoComplete="off"/>
            </Form.Item>
            <Form.Item field="strategyCount" label="规则数量">
                <Select
                    allowClear
                    placeholder="规则数量范围"
                    style={{width: defaultWidth}}
                    options={[
                        {label: "全部", value: 0},
                        ...strategyCounts.map((count) => {
                            return {label: `大于${count}`, value: count}
                        })
                    ]}
                />
            </Form.Item>
            <Form.Item label="时间" field="time">
                <DatePicker.RangePicker format="YYYY-MM-DD HH:mm:ss" showTime style={{width: defaultWidth}}/>
            </Form.Item>
            <Form.Item field="status" label="状态" defaultValue={0}>
                <Radio.Group type="button" style={{width: defaultWidth}} options={[
                    {label: "全部", value: StatusMap[Status.Status_NONE].number},
                    {label: "启用", value: StatusMap[Status.Status_ENABLE].number},
                    {label: "禁用", value: StatusMap[Status.Status_DISABLE].number},
                ]}/>
            </Form.Item>
        </Form>
    </div>
}

export default SearchForm;

