import { FC } from 'react'
import { Spin } from 'antd'

import styles from './style/index.module.less'

const Loading: FC = () => {
    return <Spin size="large" className={styles.loadingContent} />
}

export default Loading
