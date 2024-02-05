import { useContext, FC, ReactNode } from 'react'

import type { FormInstance } from 'antd'
import type { SizeType } from 'antd/es/config-provider/SizeContext'
import type { SegmentedValue } from 'antd/es/segmented'
import { Row, Col, Button, Segmented } from 'antd'

import { GlobalContext } from '@/context'

import styles from '../style/data.module.less'
import { ActionKey } from '@/apis/data'
import { ClearOutlined } from '@ant-design/icons'

export type DataOptionItem = {
    label: ReactNode
    key: string
}

export type DataOptionProps = {
    queryForm?: FormInstance
    leftOptions?: DataOptionItem[]
    rightOptions?: DataOptionItem[]
    action?: (key: ActionKey) => void
    showAdd?: boolean
    showClear?: boolean
    showSegmented?: boolean
}

const DataOption: FC<DataOptionProps> = (props) => {
    const { size, setSize } = useContext(GlobalContext)
    const {
        queryForm,
        leftOptions,
        rightOptions,
        action,
        showAdd = true,
        showClear = true,
        showSegmented = true
    } = props

    const handleSizeChange = (sizeVal: SegmentedValue) => {
        setSize?.(sizeVal as SizeType)
    }

    const clearQueryForm = () => {
        queryForm?.resetFields()
        action?.(ActionKey.RESET)
    }

    return (
        <Row className={styles.Row}>
            <Col span={8} className={styles.LeftCol}>
                {showAdd && (
                    <Button
                        type="primary"
                        onClick={() => action?.(ActionKey.ADD)}
                    >
                        添加
                    </Button>
                )}
                {leftOptions?.map(({ key, label }, index) => {
                    return (
                        <div
                            onClick={() => action?.(key as ActionKey)}
                            key={index}
                        >
                            {label}
                        </div>
                    )
                })}
            </Col>
            <Col
                span={16}
                className={styles.RightCol}
                style={{ alignItems: 'center' }}
            >
                {rightOptions?.map(({ key, label }, index) => {
                    return (
                        <div
                            onClick={() => action?.(key as ActionKey)}
                            key={index}
                        >
                            {label}
                        </div>
                    )
                })}
                {showClear && (
                    <Button
                        type="dashed"
                        onClick={clearQueryForm}
                        icon={<ClearOutlined />}
                    />
                )}

                {showSegmented && (
                    <Segmented
                        onChange={handleSizeChange}
                        value={size}
                        options={[
                            {
                                label: '大',
                                value: 'large'
                            },
                            {
                                label: '中',
                                value: 'middle'
                            },
                            {
                                label: '小',
                                value: 'small'
                            }
                        ]}
                    />
                )}
            </Col>
        </Row>
    )
}

export default DataOption
