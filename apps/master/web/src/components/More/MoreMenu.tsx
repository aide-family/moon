import React from "react";
import {Button, Dropdown, Menu} from "@arco-design/web-react";
import {IconMore} from "@arco-design/web-react/icon";

export type MoreMenuOption = {
    key: string
    label: React.ReactNode
    onClick?: () => void
}

export interface MoreMenuProps {
    children?: React.ReactNode
    options?: MoreMenuOption[]
}

export interface DropListProps {
    options: MoreMenuOption[]
    onClickMenuItem: (key: string) => void
}


const DropList: React.FC<DropListProps> = (p) => {
    const {options, onClickMenuItem} = p;

    return <Menu onClickMenuItem={onClickMenuItem} style={{margin: 0, padding: 0}}>
        {
            options.map((item) => {
                return <Menu.Item key={item.key}>{item.label}</Menu.Item>
            })
        }
    </Menu>
}

const MoreMenu: React.FC<MoreMenuProps> = (props) => {
    const {children, options = []} = props;

    const handleMenuItemClick = (key: string) => {
        const option = options.find((item) => item.key === key);
        option?.onClick?.()
    }


    return <Dropdown
        droplist={<DropList onClickMenuItem={handleMenuItemClick} options={options}/>}
        position="bottom"
    >
        {children || <Button type="text"><IconMore/></Button>}
    </Dropdown>
}

export default MoreMenu