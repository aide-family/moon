import { FC } from 'react'

import { HeightLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'

const Home: FC = () => {
    return (
        <div className="bodyContent">
            <RouteBreadcrumb />
            <HeightLine />
            <div>
                <h1>home</h1>
            </div>
        </div>
    )
}

export default Home
