import React from "react";
import AddGroup from "@/pages/strategy/group/child/AddGroup";
import groupStyle from "@/pages/strategy/group/style/group.module.less";

export interface OptionLineProps {

}

const OptionLine: React.FC<OptionLineProps> = (props) => {
    const {} = props;

    return <>
        <div className={groupStyle.OptionLineDiv}>
            <AddGroup/>
        </div>
    </>
}

export default OptionLine;