import React from 'react'

import { Button, Form } from 'antd'
import { useNavigate } from 'react-router-dom'
import { type DataFormItem, DataOption, SearchForm } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { HeightLine, PaddingLine } from '@/components/HeightLine'

const defaultPadding = 12

const Menu: React.FC = () => {
    const nav = useNavigate()

    const operationRef = React.useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()
    const searchItems: DataFormItem[] = []

    return (
        <div>
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
            <div ref={operationRef}>
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
