import React from "react";
import {Button} from "@arco-design/web-react";

export interface AddGroupProps {
    children?: React.ReactNode
}

const defaultAddGroupName = "新建分组";

const AddGroup: React.FC<AddGroupProps> = (props) => {
    const {children = defaultAddGroupName} = props;

    return <>
        <Button type="primary">
            {children}
        </Button>
    </>
}

export default AddGroup;