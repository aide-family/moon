import React from 'react'

import type { ColumnType, ColumnGroupType } from 'antd/es/table'
import type { MenuItemType } from './type'

import { Button, Form } from 'antd'
import { useNavigate } from 'react-router-dom'
import { SearchForm, type DataFormItem } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'

const defaultPadding = 12

const Menu: React.FC = () => {
    const nav = useNavigate()

    const oprationRef = React.useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()
    const searchItems: DataFormItem[] = []

    const columns: (
        | ColumnGroupType<MenuItemType>
        | ColumnType<MenuItemType>
    )[] = []
    const [dataSource, setDataSource] = React.useState<MenuItemType[]>([])

    return (
        <div className="bodyContent">
            <Button
                onClick={() => {
                    nav('/home', {
                        state: {
                            name: 'hello'
                        }
                    })
                }}
            >
                hoem
            </Button>
            <div ref={oprationRef}>
                <RouteBreadcrumb />
                <SearchForm form={queryForm} items={searchItems} />
                <HeightLine />
                <DataOption queryForm={queryForm} />
                <PaddingLine
                    padding={defaultPadding}
                    height={1}
                    borderRadius={4}
                />
                B
            </div>
        </div>
    )
}

export default Menu
