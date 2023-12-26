import {DataOptionItem} from "@/components/Data/DataOption/DataOption.tsx";
import {ActionKey} from "@/apis/data.ts";
import {Button} from "antd";
import {DataFormItem} from "@/components/Data";

const leftOptions: DataOptionItem[] = [
    // {
    //     key: ActionKey.BATCH_IMPORT,
    //     label: (
    //         <Button type="primary" loading={loading}>
    //             批量导入
    //         </Button>
    //     )
    // }
]

const rightOptions = (loading?: boolean): DataOptionItem[] => [
    {
        key: ActionKey.REFRESH,
        label: (
            <Button
                type="primary"
                loading={loading}
            >
                刷新
            </Button>
        )
    }
]

const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '关键词'
    },
    {
        name: 'status',
        label: '状态'
    }
]

export const options = {
    leftOptions,
    rightOptions,
    searchItems
}