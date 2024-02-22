import { FC } from 'react'

import { Carousel } from 'antd'
import Banner from '@/assets/buner.png'
import Logo from '@/assets/logo.svg'
import styles from '../style/login.module.less'

export type CarouselItemProps = {
    title: string
    subTitle: string
    img: string
}

const Title = () => {
    return (
        <div className={styles.LoginCarouselLogo}>
            <img alt="" src={Logo} className={styles.LoginCarouselLogoImg} />
            <div>Moon</div>
        </div>
    )
}

const CarouselItem: FC<CarouselItemProps> = (props) => {
    const { title, subTitle, img } = props
    return (
        <div className={styles.ContentItem}>
            <div className={styles.ContentItemTitle}>{title}</div>
            <div className={styles.ContentItemSubTitle}>{subTitle}</div>
            <img alt={title} src={img} className={styles.ContentItemImage} />
        </div>
    )
}

const LoginCarousel: FC = () => {
    const items: CarouselItemProps[] = [
        {
            title: '内置了常见问题的解决方案',
            subTitle: '设备管理，节点管理等应有尽有',
            img: Banner
        },
        {
            title: '接入可视化增强工具Grafana',
            subTitle: '实现灵活的区块式组合',
            img: Banner
        },
        {
            title: '支持prometheus监控',
            subTitle: '告警规则，规则组管理, 告警通知等',
            img: Banner
        }
    ]

    return (
        <div className={styles.LoginCarousel}>
            <Title />
            <Carousel autoplay className={styles.Content}>
                {items.map((item, index) => (
                    <CarouselItem
                        key={index}
                        title={item.title}
                        subTitle={item.subTitle}
                        img={item.img}
                    />
                ))}
            </Carousel>
        </div>
    )
}

export default LoginCarousel
