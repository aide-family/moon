import { DataOption, SearchForm } from '@/components/Data'
import { HeightLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Form } from 'antd'
import React, { useState } from 'react'
import { leftOptions, rightOptions, searchItems } from './options'
import { ActionKey } from '@/apis/data'
import { EditStrategyModal } from './child/EditStrategyModal'

let timer: NodeJS.Timeout

const Strategy: React.FC = () => {
    const [queryForm] = Form.useForm()

    const [openEditModal, setOpenEditModal] = useState(false)

    const handleCancelStrategyEditModal = () => {
        setOpenEditModal(false)
    }

    const handleStrategyEditModalOk = () => {
        setOpenEditModal(false)
    }

    const openStrategyEditModal = () => {
        setOpenEditModal(true)
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (
        changedValues: any, // TODO 不要any
        allValues: any
    ) => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            console.log(changedValues, allValues)
        }, 500)
    }

    const handleOptionClick = (val: ActionKey) => {
        switch (val) {
            case ActionKey.ADD:
                openStrategyEditModal()
                break
        }
    }

    return (
        <div id="strategy-content" className="bodyContent">
            <EditStrategyModal
                open={openEditModal}
                onCancel={handleCancelStrategyEditModal}
                onOk={handleStrategyEditModalOk}
            />
            <div id="strategy-content-op">
                <RouteBreadcrumb />
                <HeightLine />
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearFormValuesChange
                    }}
                />
                <HeightLine />
                <DataOption
                    queryForm={queryForm}
                    rightOptions={rightOptions}
                    leftOptions={leftOptions}
                    action={handleOptionClick}
                />
            </div>
        </div>
    )
}

export default Strategy
