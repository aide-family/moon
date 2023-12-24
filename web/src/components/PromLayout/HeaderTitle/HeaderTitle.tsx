import { FC, ReactNode } from 'react'

import logoIcon from '@/assets/logo.svg'

import styles from './style/index.module.less'

export type HeaderTitleProps = {
    title?: ReactNode
}

const HeaderTitle: FC<HeaderTitleProps> = (props) => {
    const { title = 'AideDevOps' } = props

    return (
        <div className={styles.headerTitle}>
            <img src={logoIcon} alt="" style={{ height: 32, width: 32 }} />
            {title}
        </div>
    )
}

export default HeaderTitle
